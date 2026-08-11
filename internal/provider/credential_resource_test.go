package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

func TestCredentialResourceSchemaGitHubApp(t *testing.T) {
	var resp resource.SchemaResponse
	(&CredentialResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())

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
