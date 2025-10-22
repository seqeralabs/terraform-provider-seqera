package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = sshCredentialValidator{}

type sshCredentialValidator struct{}

func (v sshCredentialValidator) Description(_ context.Context) string {
	return "validates that the credential provider is set to 'ssh'"
}

func (v sshCredentialValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v sshCredentialValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	provider := req.ConfigValue.ValueString()

	if provider != "ssh" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Credential Provider",
			fmt.Sprintf("Credential provider must be 'ssh'. Currently only SSH credentials are supported for managed identities. Got: '%s'", provider),
		)
	}
}

func SSHCredentialValidator() validator.String {
	return sshCredentialValidator{}
}
