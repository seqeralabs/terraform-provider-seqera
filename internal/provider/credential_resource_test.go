package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestCredentialResourceSchemaGitHubApp(t *testing.T) {
	var resp resource.SchemaResponse
	(&CredentialResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	providerType, ok := resp.Schema.Attributes["provider_type"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, providerType.Validators)
	var validation validator.StringResponse
	providerType.Validators[0].ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("provider_type"),
		ConfigValue: types.StringValue("github_app"),
	}, &validation)
	require.False(t, validation.Diagnostics.HasError(), "provider_type must accept github_app")

	keys, ok := resp.Schema.Attributes["keys"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	githubApp, ok := keys.Attributes["github_app"].(schema.SingleNestedAttribute)
	require.True(t, ok)

	for _, field := range []string{"client_secret", "private_key", "webhook_secret"} {
		attribute, ok := githubApp.Attributes[field].(schema.StringAttribute)
		require.True(t, ok, "keys.github_app.%s should be a string attribute", field)
		require.True(t, attribute.Sensitive, "keys.github_app.%s should be sensitive", field)
	}
}
