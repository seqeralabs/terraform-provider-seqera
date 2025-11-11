package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = AWSCredentialKeysValidatorValidator{}

type AWSCredentialKeysValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v AWSCredentialKeysValidatorValidator) Description(_ context.Context) string {
	return "validates that either (access_key and secret_key) or assume_role_arn must be provided"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v AWSCredentialKeysValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v AWSCredentialKeysValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Get the sibling field values
	var accessKeyValue types.String
	var secretKeyValue types.String
	var assumeRoleArnValue types.String

	accessKeyPath := req.Path.ParentPath().AtName("access_key")
	secretKeyPath := req.Path.ParentPath().AtName("secret_key")
	assumeRoleArnPath := req.Path.ParentPath().AtName("assume_role_arn")

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, accessKeyPath, &accessKeyValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, secretKeyPath, &secretKeyValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, assumeRoleArnPath, &assumeRoleArnValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase
	if accessKeyValue.IsUnknown() || secretKeyValue.IsUnknown() || assumeRoleArnValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	// Check if each field is provided
	accessKeyProvided := !accessKeyValue.IsNull() && accessKeyValue.ValueString() != ""
	secretKeyProvided := !secretKeyValue.IsNull() && secretKeyValue.ValueString() != ""
	assumeRoleArnProvided := !assumeRoleArnValue.IsNull() && assumeRoleArnValue.ValueString() != ""

	// Rule 1: If access_key is provided, secret_key must also be provided
	if accessKeyProvided && !secretKeyProvided {
		resp.Diagnostics.AddAttributeError(
			secretKeyPath,
			"Missing Required Attribute",
			"The 'secret_key' attribute is required when 'access_key' is provided. AWS credentials require both access_key and secret_key together.",
		)
		return
	}

	// Rule 2: If secret_key is provided, access_key must also be provided
	if secretKeyProvided && !accessKeyProvided {
		resp.Diagnostics.AddAttributeError(
			accessKeyPath,
			"Missing Required Attribute",
			"The 'access_key' attribute is required when 'secret_key' is provided. AWS credentials require both access_key and secret_key together.",
		)
		return
	}

	// Rule 3: At least one authentication method must be provided
	if !accessKeyProvided && !secretKeyProvided && !assumeRoleArnProvided {
		// Add error to the current field being validated
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Configuration",
			"AWS credentials require either 'assume_role_arn' or both 'access_key' and 'secret_key' to be provided. At least one authentication method must be configured.",
		)
		return
	}
}

func AWSCredentialKeysValidator() validator.String {
	return AWSCredentialKeysValidatorValidator{}
}
