package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AuthValidatingProvider wraps the generated SeqeraProvider and checks that
// exactly one authentication method is configured before the generated
// Configure builds the API client.
//
// With two alternative security schemes (BearerAuth and ClientCredentials),
// Speakeasy no longer generates the "Missing Provider Security Configuration"
// check, and it silently ignores half a client credentials pair and lets
// client credentials win over a personal access token. The check lives here,
// rather than in provider.go, because provider.go is regenerated. main.go
// serves this wrapper and is listed in .genignore.
//
// The check runs in Configure, as the generated one did, so `terraform
// validate` keeps working without credentials.
type AuthValidatingProvider struct {
	*SeqeraProvider
}

var _ provider.Provider = (*AuthValidatingProvider)(nil)
var _ provider.ProviderWithActions = (*AuthValidatingProvider)(nil)
var _ provider.ProviderWithEphemeralResources = (*AuthValidatingProvider)(nil)
var _ provider.ProviderWithFunctions = (*AuthValidatingProvider)(nil)

// NewWithAuthValidation is New with the authentication check.
func NewWithAuthValidation(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AuthValidatingProvider{SeqeraProvider: &SeqeraProvider{version: version}}
	}
}

func (p *AuthValidatingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SeqeraProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateAuthConfig(data, os.Getenv)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.SeqeraProvider.Configure(ctx, req, resp)
}

// authSetting is one credential, settable through a provider attribute or
// its environment variable. The provider attribute takes precedence, as in
// the generated Configure.
type authSetting struct {
	attribute string
	envVar    string
	value     types.String
}

// resolve reports where the setting's effective value comes from, or "" when
// it is unset.
func (s authSetting) resolve(getenv func(string) string) string {
	if !s.value.IsNull() && !s.value.IsUnknown() && s.value.ValueString() != "" {
		return fmt.Sprintf("%s in the provider block", s.attribute)
	}
	if getenv(s.envVar) != "" {
		return fmt.Sprintf("the %s environment variable", s.envVar)
	}
	return ""
}

func validateAuthConfig(data SeqeraProviderModel, getenv func(string) string) diag.Diagnostics {
	var diags diag.Diagnostics

	bearer := authSetting{"bearer_auth", "TOWER_ACCESS_TOKEN", data.BearerAuth}
	clientID := authSetting{"client_id", "TOWER_CLIENT_ID", data.ClientID}
	clientSecret := authSetting{"client_secret", "TOWER_CLIENT_SECRET", data.ClientSecret}

	// A value computed from other resources is unknown until apply; it cannot
	// be checked yet, and the generated Configure skips it too.
	for _, s := range []authSetting{bearer, clientID, clientSecret} {
		if s.value.IsUnknown() {
			return diags
		}
	}

	bearerSource := bearer.resolve(getenv)
	idSource := clientID.resolve(getenv)
	secretSource := clientSecret.resolve(getenv)

	var ccSources []string
	for _, src := range []string{idSource, secretSource} {
		if src != "" {
			ccSources = append(ccSources, src)
		}
	}

	switch {
	case bearerSource == "" && len(ccSources) == 0:
		diags.AddError(
			"Missing Provider Authentication",
			"Configure either a personal access token (the bearer_auth attribute or the TOWER_ACCESS_TOKEN environment variable) "+
				"or service account client credentials (the client_id and client_secret attributes, "+
				"or the TOWER_CLIENT_ID and TOWER_CLIENT_SECRET environment variables).",
		)
		return diags
	case bearerSource != "" && len(ccSources) > 0:
		diags.AddError(
			"Conflicting Provider Authentication",
			fmt.Sprintf(
				"Both a personal access token (%s) and service account client credentials (%s) are configured. "+
					"Configure only one of them.",
				bearerSource, strings.Join(ccSources, ", "),
			),
		)
	}

	if len(ccSources) == 1 {
		missing := "client_secret attribute or the TOWER_CLIENT_SECRET environment variable"
		if idSource == "" {
			missing = "client_id attribute or the TOWER_CLIENT_ID environment variable"
		}
		diags.AddError(
			"Incomplete Service Account Client Credentials",
			fmt.Sprintf(
				"Client credentials are partially configured: %s is set, but the %s is not. "+
					"Set both the client ID and the client secret to authenticate as a service account.",
				ccSources[0], missing,
			),
		)
	}

	return diags
}
