package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringAzureCredentialEntraValidator{}

type StringAzureCredentialEntraValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringAzureCredentialEntraValidator) Description(_ context.Context) string {
	return "validates that tenant_id, client_id, and client_secret are required when using Entra/Cloud authentication"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringAzureCredentialEntraValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringAzureCredentialEntraValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Get the discriminator field value
	var discriminatorValue types.String
	discriminatorPath := req.Path.ParentPath().AtName("discriminator")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, discriminatorPath, &discriminatorValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown discriminator during plan phase
	if discriminatorValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	// Determine which field we're validating
	fieldName := req.Path.Steps()[len(req.Path.Steps())-1].String()

	// If discriminator is NOT "azure" (i.e., it's entra or cloud), these fields are required
	isEntraOrCloud := !discriminatorValue.IsNull() && discriminatorValue.ValueString() != "azure"
	fieldIsEmpty := req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == ""

	if isEntraOrCloud && fieldIsEmpty {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Attribute",
			"The '"+fieldName+"' attribute is required when using Entra or Cloud authentication. "+
				"For shared key authentication, use batch_key and storage_key instead.",
		)
		return
	}

	// If discriminator IS "azure" (shared key), these fields should not be set
	if !isEntraOrCloud && !fieldIsEmpty {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Azure Credential Configuration",
			"The '"+fieldName+"' field can only be used with Entra or Cloud authentication. "+
				"For shared key authentication (discriminator='azure'), use batch_key and storage_key instead.",
		)
	}
}

func AzureCredentialEntraValidator() validator.String {
	return StringAzureCredentialEntraValidator{}
}
