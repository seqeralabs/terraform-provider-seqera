package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = GoogleCredentialKeysValidatorValidator{}

type GoogleCredentialKeysValidatorValidator struct{}

func (v GoogleCredentialKeysValidatorValidator) Description(_ context.Context) string {
	return "validates that either 'data' (service account key JSON) or both 'workload_identity_provider' and 'service_account_email' must be provided"
}

func (v GoogleCredentialKeysValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v GoogleCredentialKeysValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	var dataValue types.String
	var wifProviderValue types.String
	var saEmailValue types.String

	dataPath := req.Path.ParentPath().AtName("data")
	wifProviderPath := req.Path.ParentPath().AtName("workload_identity_provider")
	saEmailPath := req.Path.ParentPath().AtName("service_account_email")

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, dataPath, &dataValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, wifProviderPath, &wifProviderValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, saEmailPath, &saEmailValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if dataValue.IsUnknown() || wifProviderValue.IsUnknown() || saEmailValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	dataProvided := !dataValue.IsNull() && dataValue.ValueString() != ""
	wifProviderProvided := !wifProviderValue.IsNull() && wifProviderValue.ValueString() != ""
	saEmailProvided := !saEmailValue.IsNull() && saEmailValue.ValueString() != ""

	if dataProvided && (wifProviderProvided || saEmailProvided) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting Authentication Methods",
			"Google credentials must use either 'data' (service account key JSON) or Workload Identity Federation ('workload_identity_provider' + 'service_account_email'), not both. Remove one to disambiguate.",
		)
		return
	}

	if wifProviderProvided && !saEmailProvided {
		resp.Diagnostics.AddAttributeError(
			saEmailPath,
			"Missing Required Attribute",
			"The 'service_account_email' attribute is required when 'workload_identity_provider' is provided. Workload Identity Federation needs both to identify the service account Seqera should impersonate.",
		)
		return
	}

	if saEmailProvided && !wifProviderProvided {
		resp.Diagnostics.AddAttributeError(
			wifProviderPath,
			"Missing Required Attribute",
			"The 'workload_identity_provider' attribute is required when 'service_account_email' is provided. Workload Identity Federation needs both to identify the pool provider that trusts Seqera as an OIDC issuer.",
		)
		return
	}

	if !dataProvided && !wifProviderProvided && !saEmailProvided {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Configuration",
			"Google credentials require either 'data' (service account key JSON) or Workload Identity Federation ('workload_identity_provider' + 'service_account_email'). At least one authentication method must be configured.",
		)
		return
	}
}

func GoogleCredentialKeysValidator() validator.String {
	return GoogleCredentialKeysValidatorValidator{}
}
