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
	// Allow unknown values during plan phase
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Determine which field we're validating
	fieldName := req.Path.Steps()[len(req.Path.Steps())-1].String()
	fieldIsEmpty := req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == ""

	// Get the values of all Entra/Cloud authentication fields
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

	// Check if we're in Entra/Cloud mode (at least one Entra field is set)
	isEntraOrCloudMode := (!tenantID.IsNull() && !tenantID.IsUnknown() && tenantID.ValueString() != "") ||
		(!clientID.IsNull() && !clientID.IsUnknown() && clientID.ValueString() != "") ||
		(!clientSecret.IsNull() && !clientSecret.IsUnknown() && clientSecret.ValueString() != "")

	// If in Entra/Cloud mode, all three fields (tenant_id, client_id, client_secret) must be set
	if isEntraOrCloudMode && fieldIsEmpty {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Attribute",
			"The '"+fieldName+"' attribute is required when using Entra or Cloud authentication. "+
				"All three fields (tenant_id, client_id, client_secret) must be provided together. "+
				"For shared key authentication, use batch_key and storage_key instead.",
		)
		return
	}

	// Get the values of shared key fields to check if they're set
	var batchKey types.String
	var storageKey types.String

	batchKeyPath := req.Path.ParentPath().AtName("batch_key")
	storageKeyPath := req.Path.ParentPath().AtName("storage_key")

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, batchKeyPath, &batchKey)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, storageKeyPath, &storageKey)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if shared key fields are set
	isSharedKeyMode := (!batchKey.IsNull() && !batchKey.IsUnknown() && batchKey.ValueString() != "") ||
		(!storageKey.IsNull() && !storageKey.IsUnknown() && storageKey.ValueString() != "")

	// If shared key mode is detected and Entra/Cloud field is set, that's an error
	if isSharedKeyMode && !fieldIsEmpty {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Azure Credential Configuration",
			"The '"+fieldName+"' field can only be used with Entra or Cloud authentication. "+
				"For shared key authentication, use batch_key and storage_key instead. "+
				"Do not mix shared key fields (batch_key, storage_key) with Entra/Cloud fields (tenant_id, client_id, client_secret).",
		)
	}
}

func AzureCredentialEntraValidator() validator.String {
	return StringAzureCredentialEntraValidator{}
}
