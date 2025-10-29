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

	// Determine authentication mode by checking if Entra/Cloud fields are set
	// If tenant_id, client_id, or client_secret are set, we're in Entra/Cloud mode
	var tenantID types.String
	var clientID types.String
	var clientSecret types.String

	tenantIDPath := req.Path.ParentPath().AtName("tenant_id")
	clientIDPath := req.Path.ParentPath().AtName("client_id")
	clientSecretPath := req.Path.ParentPath().AtName("client_secret")

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, tenantIDPath, &tenantID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, clientIDPath, &clientID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, clientSecretPath, &clientSecret)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine which field we're validating
	fieldName := req.Path.Steps()[len(req.Path.Steps())-1].String()

	// Check if any Entra/Cloud fields are set (indicating Entra/Cloud authentication mode)
	isEntraOrCloudMode := (!tenantID.IsNull() && !tenantID.IsUnknown()) ||
		(!clientID.IsNull() && !clientID.IsUnknown()) ||
		(!clientSecret.IsNull() && !clientSecret.IsUnknown())

	// If Entra/Cloud mode is detected, shared key fields should not be set
	if isEntraOrCloudMode {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Azure Credential Configuration",
			"The '"+fieldName+"' field can only be used with shared key authentication. "+
				"For Entra or Cloud authentication, use tenant_id, client_id, and client_secret instead. "+
				"Do not mix shared key fields (batch_key, storage_key) with Entra/Cloud fields (tenant_id, client_id, client_secret).",
		)
	}
}

func AzureCredentialSharedKeyValidator() validator.String {
	return StringAzureCredentialSharedKeyValidator{}
}
