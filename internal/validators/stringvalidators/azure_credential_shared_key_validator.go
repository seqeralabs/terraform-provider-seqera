package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringAzureCredentialSharedKeyValidator{}

type StringAzureCredentialSharedKeyValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringAzureCredentialSharedKeyValidator) Description(_ context.Context) string {
	return "validates that batch_key and storage_key can only be set when using shared key authentication (discriminator='azure')"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringAzureCredentialSharedKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringAzureCredentialSharedKeyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip if value is null or unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Get the discriminator field value
	var discriminatorValue types.String
	discriminatorPath := req.Path.ParentPath().AtName("discriminator")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, discriminatorPath, &discriminatorValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown discriminator during plan phase
	if discriminatorValue.IsUnknown() {
		return
	}

	// Determine which field we're validating
	fieldName := req.Path.Steps()[len(req.Path.Steps())-1].String()

	// If discriminator is not "azure" (shared key mode), these fields should not be set
	if !discriminatorValue.IsNull() && discriminatorValue.ValueString() != "azure" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Azure Credential Configuration",
			"The '"+fieldName+"' field can only be used with shared key authentication (discriminator='azure'). "+
				"For Entra or Cloud authentication, use tenant_id, client_id, and client_secret instead.",
		)
	}
}

func AzureCredentialSharedKeyValidator() validator.String {
	return StringAzureCredentialSharedKeyValidator{}
}
