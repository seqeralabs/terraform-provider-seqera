package stringvalidators

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringKeyVaultURLValidatorValidator{}

// keyVaultURLPattern is the RE2-compatible part of the upstream OpenAPI pattern
//
//	^https://[a-zA-Z](?!.*--)[a-zA-Z0-9-]{1,22}[a-zA-Z0-9]\.vault\.azure\.net/?$
//
// Go's regexp does not support the negative lookahead, so Speakeasy drops the
// pattern entirely; the no-consecutive-hyphens rule is checked separately.
var keyVaultURLPattern = regexp.MustCompile(`^https://([a-zA-Z][a-zA-Z0-9-]{1,22}[a-zA-Z0-9])\.vault\.azure\.net/?$`)

type StringKeyVaultURLValidatorValidator struct{}

func (v StringKeyVaultURLValidatorValidator) Description(_ context.Context) string {
	return "validates that key_vault_url is in the form https://<vault-name>.vault.azure.net"
}

func (v StringKeyVaultURLValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringKeyVaultURLValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	url := req.ConfigValue.ValueString()
	match := keyVaultURLPattern.FindStringSubmatch(url)
	if match != nil && !strings.Contains(match[1], "--") {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Key Vault URL",
		fmt.Sprintf(
			"key_vault_url must be in the form https://<vault-name>.vault.azure.net. "+
				"The vault name must be 3-24 characters long, start with a letter, end with a letter or digit, "+
				"and contain only letters, digits and non-consecutive hyphens. Got: %q", url,
		),
	)
}

func KeyVaultURLValidator() validator.String {
	return StringKeyVaultURLValidatorValidator{}
}
