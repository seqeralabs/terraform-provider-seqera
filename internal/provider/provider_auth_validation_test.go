package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestValidateAuthConfig(t *testing.T) {
	str := types.StringValue
	null := types.StringNull()
	unknown := types.StringUnknown()

	tests := []struct {
		name      string
		bearer    types.String
		id        types.String
		secret    types.String
		env       map[string]string
		wantError []string // summaries, in order; empty means valid
		wantInMsg string   // substring of the first error's detail
	}{
		// Valid
		{name: "PAT in provider block", bearer: str("pat"), id: null, secret: null},
		{name: "PAT in env", bearer: null, id: null, secret: null, env: map[string]string{"TOWER_ACCESS_TOKEN": "pat"}},
		{name: "client credentials in provider block", bearer: null, id: str("sa_1"), secret: str("s")},
		{name: "client credentials in env", bearer: null, id: null, secret: null,
			env: map[string]string{"TOWER_CLIENT_ID": "sa_1", "TOWER_CLIENT_SECRET": "s"}},
		{name: "client credentials mixed sources", bearer: null, id: str("sa_1"), secret: null,
			env: map[string]string{"TOWER_CLIENT_SECRET": "s"}},
		{name: "token URL alone does not count", bearer: str("pat"), id: null, secret: null,
			env: map[string]string{"TOWER_TOKEN_URL": "https://h/api/token"}},
		{name: "unknown value is not checked", bearer: unknown, id: null, secret: null},
		{name: "empty attribute falls back to env", bearer: null, id: str(""), secret: str(""),
			env: map[string]string{"TOWER_ACCESS_TOKEN": "pat"}},

		// Missing
		{name: "nothing set", bearer: null, id: null, secret: null,
			wantError: []string{"Missing Provider Authentication"}, wantInMsg: "TOWER_CLIENT_ID"},
		{name: "empty values only", bearer: str(""), id: str(""), secret: str(""),
			wantError: []string{"Missing Provider Authentication"}},

		// Partial
		{name: "client ID only", bearer: null, id: str("sa_1"), secret: null,
			wantError: []string{"Incomplete Service Account Client Credentials"},
			wantInMsg: "client_id in the provider block is set, but the client_secret attribute or the TOWER_CLIENT_SECRET environment variable is not"},
		{name: "client secret only, from env", bearer: null, id: null, secret: null,
			env:       map[string]string{"TOWER_CLIENT_SECRET": "s"},
			wantError: []string{"Incomplete Service Account Client Credentials"},
			wantInMsg: "the TOWER_CLIENT_SECRET environment variable is set, but the client_id attribute or the TOWER_CLIENT_ID environment variable is not"},

		// Conflicting
		{name: "PAT and client credentials", bearer: str("pat"), id: str("sa_1"), secret: str("s"),
			wantError: []string{"Conflicting Provider Authentication"},
			wantInMsg: "(bearer_auth in the provider block) and service account client credentials (client_id in the provider block, client_secret in the provider block)"},
		{name: "PAT in env and client credentials in block", bearer: null, id: str("sa_1"), secret: str("s"),
			env:       map[string]string{"TOWER_ACCESS_TOKEN": "pat"},
			wantError: []string{"Conflicting Provider Authentication"},
			wantInMsg: "the TOWER_ACCESS_TOKEN environment variable"},
		{name: "PAT and partial client credentials", bearer: str("pat"), id: str("sa_1"), secret: null,
			wantError: []string{"Conflicting Provider Authentication", "Incomplete Service Account Client Credentials"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := SeqeraProviderModel{
				BearerAuth:   tt.bearer,
				ClientID:     tt.id,
				ClientSecret: tt.secret,
				ServerURL:    types.StringNull(),
				TokenURL:     types.StringNull(),
			}
			getenv := func(k string) string { return tt.env[k] }

			diags := validateAuthConfig(data, getenv)

			var got []string
			for _, d := range diags.Errors() {
				got = append(got, d.Summary())
			}
			if strings.Join(got, "|") != strings.Join(tt.wantError, "|") {
				t.Fatalf("got errors %q, want %q (%v)", got, tt.wantError, diags)
			}
			if tt.wantInMsg != "" && !strings.Contains(diags.Errors()[0].Detail(), tt.wantInMsg) {
				t.Fatalf("got detail %q, want it to contain %q", diags.Errors()[0].Detail(), tt.wantInMsg)
			}
		})
	}
}

// Runs the wrapper's Configure with a real provider config: an invalid
// configuration must stop before the generated Configure creates a client,
// and a valid one must reach it.
func TestAuthValidatingProviderConfigure(t *testing.T) {
	for _, k := range []string{"TOWER_ACCESS_TOKEN", "TOWER_CLIENT_ID", "TOWER_CLIENT_SECRET", "TOWER_TOKEN_URL"} {
		t.Setenv(k, "")
	}

	tests := []struct {
		name       string
		attrs      map[string]string
		wantClient bool
	}{
		{"valid PAT", map[string]string{"bearer_auth": "pat"}, true},
		{"valid client credentials", map[string]string{"client_id": "sa_1", "client_secret": "s"}, true},
		{"nothing set", nil, false},
		{"both set", map[string]string{"bearer_auth": "pat", "client_id": "sa_1", "client_secret": "s"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			p := NewWithAuthValidation("test")()

			var schemaResp provider.SchemaResponse
			p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)

			objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
			vals := map[string]tftypes.Value{}
			for name := range objType.AttributeTypes {
				var v any
				if s, ok := tt.attrs[name]; ok {
					v = s
				}
				vals[name] = tftypes.NewValue(tftypes.String, v)
			}

			req := provider.ConfigureRequest{Config: tfsdk.Config{
				Schema: schemaResp.Schema,
				Raw:    tftypes.NewValue(objType, vals),
			}}
			var resp provider.ConfigureResponse
			p.Configure(ctx, req, &resp)

			if gotClient := resp.ResourceData != nil; gotClient != tt.wantClient {
				t.Fatalf("got client %v, want %v (diags: %v)", gotClient, tt.wantClient, resp.Diagnostics)
			}
			if resp.Diagnostics.HasError() == tt.wantClient {
				t.Fatalf("got diagnostics %v, want errors: %v", resp.Diagnostics, !tt.wantClient)
			}
		})
	}
}
