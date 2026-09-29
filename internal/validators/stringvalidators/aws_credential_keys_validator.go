package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = AWSCredentialKeysValidatorValidator{}

type AWSCredentialKeysValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v AWSCredentialKeysValidatorValidator) Description(_ context.Context) string {
	return "validates that either (access_key and secret_key) or assume_role_arn must be provided, and that the fields match the selected mode ('role' and 'workloadIdentity' require assume_role_arn without static keys)"
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
	var modeValue types.String
	var useExternalIDValue types.Bool

	accessKeyPath := req.Path.ParentPath().AtName("access_key")
	secretKeyPath := req.Path.ParentPath().AtName("secret_key")
	assumeRoleArnPath := req.Path.ParentPath().AtName("assume_role_arn")
	modePath := req.Path.ParentPath().AtName("mode")
	useExternalIDPath := req.Path.ParentPath().AtName("use_external_id")

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, accessKeyPath, &accessKeyValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, secretKeyPath, &secretKeyValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, assumeRoleArnPath, &assumeRoleArnValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, modePath, &modeValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, useExternalIDPath, &useExternalIDValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase
	if accessKeyValue.IsUnknown() || secretKeyValue.IsUnknown() || assumeRoleArnValue.IsUnknown() ||
		modeValue.IsUnknown() || useExternalIDValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	// Check if each field is provided
	accessKeyProvided := !accessKeyValue.IsNull() && accessKeyValue.ValueString() != ""
	secretKeyProvided := !secretKeyValue.IsNull() && secretKeyValue.ValueString() != ""
	assumeRoleArnProvided := !assumeRoleArnValue.IsNull() && assumeRoleArnValue.ValueString() != ""
	useExternalID := !useExternalIDValue.IsNull() && useExternalIDValue.ValueBool()
	mode := modeValue.ValueString()

	// Mode rules mirror the Platform's AwsSecurityKeys.validate(), so a mismatch fails at plan time
	// rather than at apply. Errors are attached to fixed paths so the three attributes carrying this
	// validator report each problem once.
	if mode == "role" || mode == "workloadIdentity" {
		if accessKeyProvided || secretKeyProvided {
			errPath := accessKeyPath
			if !accessKeyProvided {
				errPath = secretKeyPath
			}
			resp.Diagnostics.AddAttributeError(
				errPath,
				"Conflicting Attributes",
				fmt.Sprintf("The 'access_key' and 'secret_key' attributes must not be set when 'mode' is %q. This mode authenticates by assuming 'assume_role_arn' without static credentials.", mode),
			)
			return
		}
		if !assumeRoleArnProvided {
			resp.Diagnostics.AddAttributeError(
				assumeRoleArnPath,
				"Missing Required Attribute",
				fmt.Sprintf("The 'assume_role_arn' attribute is required when 'mode' is %q.", mode),
			)
			return
		}
	}

	if mode == "workloadIdentity" && useExternalID {
		resp.Diagnostics.AddAttributeError(
			useExternalIDPath,
			"Conflicting Attributes",
			"The 'use_external_id' attribute must not be true when 'mode' is \"workloadIdentity\". Workload identity federation does not use an External ID.",
		)
		return
	}

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
