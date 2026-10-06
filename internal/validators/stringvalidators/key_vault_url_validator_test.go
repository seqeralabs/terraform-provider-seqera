package stringvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestKeyVaultURLValidator(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		{name: "valid url", value: types.StringValue("https://my-vault.vault.azure.net")},
		{name: "valid url with trailing slash", value: types.StringValue("https://my-vault.vault.azure.net/")},
		{name: "minimum length name", value: types.StringValue("https://abc.vault.azure.net")},
		{name: "maximum length name", value: types.StringValue("https://a23456789012345678901234.vault.azure.net")},
		{name: "null is skipped", value: types.StringNull()},
		{name: "unknown is skipped", value: types.StringUnknown()},
		{name: "empty string", value: types.StringValue(""), expectError: true},
		{name: "http scheme", value: types.StringValue("http://my-vault.vault.azure.net"), expectError: true},
		{name: "missing scheme", value: types.StringValue("my-vault.vault.azure.net"), expectError: true},
		{name: "wrong domain", value: types.StringValue("https://my-vault.vault.azure.com"), expectError: true},
		{name: "name too short", value: types.StringValue("https://ab.vault.azure.net"), expectError: true},
		{name: "name too long", value: types.StringValue("https://a234567890123456789012345.vault.azure.net"), expectError: true},
		{name: "name starts with digit", value: types.StringValue("https://1vault.vault.azure.net"), expectError: true},
		{name: "name ends with hyphen", value: types.StringValue("https://my-vault-.vault.azure.net"), expectError: true},
		{name: "consecutive hyphens", value: types.StringValue("https://my--vault.vault.azure.net"), expectError: true},
		{name: "trailing path", value: types.StringValue("https://my-vault.vault.azure.net/secrets"), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("key_vault_url"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			KeyVaultURLValidator().ValidateString(context.Background(), req, resp)

			if got := resp.Diagnostics.HasError(); got != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v: %v", tt.expectError, got, resp.Diagnostics)
			}
		})
	}
}
