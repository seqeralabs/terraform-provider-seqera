package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func resourceSchema(t *testing.T, r resource.Resource) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func stringAttributeErrors(t *testing.T, s schema.Schema, attr, value string) bool {
	t.Helper()
	a, ok := s.Attributes[attr].(schema.StringAttribute)
	if !ok {
		t.Fatalf("%s is not a string attribute", attr)
	}
	for _, v := range a.Validators {
		var resp validator.StringResponse
		v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root(attr), ConfigValue: types.StringValue(value)}, &resp)
		if resp.Diagnostics.HasError() {
			return true
		}
	}
	return false
}

// The Agent Backend rejects an agent without a service account (a null
// serviceAccountId fails its request validation), so the provider requires one.
func TestAgentServiceAccountIsRequired(t *testing.T) {
	s := resourceSchema(t, NewAgentResource())
	a := s.Attributes["service_account_id"]
	if a == nil {
		t.Fatal("missing attribute service_account_id")
	}
	if !a.IsRequired() || a.IsComputed() {
		t.Errorf("service_account_id: required=%v computed=%v, want required and not computed", a.IsRequired(), a.IsComputed())
	}
}

// The agent PUT replaces every field, so removing one of these from the
// configuration must reach the API as null. An Optional+Computed attribute
// keeps its prior state value instead, so they must be plain Optional.
func TestAgentClearableAttributesAreNotComputed(t *testing.T) {
	s := resourceSchema(t, NewAgentResource())
	for _, name := range []string{"description", "github_app_credential_id"} {
		a := s.Attributes[name]
		if a == nil {
			t.Fatalf("missing attribute %s", name)
		}
		if !a.IsOptional() || a.IsComputed() {
			t.Errorf("%s: optional=%v computed=%v, want optional and not computed", name, a.IsOptional(), a.IsComputed())
		}
	}
}

// The backend trims description and github_app_credential_id and stores a
// blank value as null, which would make the applied value differ from the
// configured one.
func TestAgentTrimmedAttributesRejectSurroundingWhitespace(t *testing.T) {
	s := resourceSchema(t, NewAgentResource())
	for _, attr := range []string{"description", "github_app_credential_id"} {
		for _, bad := range []string{"", " padded", "padded ", " "} {
			if !stringAttributeErrors(t, s, attr, bad) {
				t.Errorf("%s = %q: expected a validation error", attr, bad)
			}
		}
		if stringAttributeErrors(t, s, attr, "Explains failed runs") {
			t.Errorf("%s: a normal value was rejected", attr)
		}
	}
}

// Service account names are validated by the backend as user names:
// 2-39 characters of lower-case letters, digits and single dashes, starting
// and ending with a letter or digit.
func TestServiceAccountNameMatchesBackendRule(t *testing.T) {
	s := resourceSchema(t, NewServiceAccountResource())
	for _, bad := range []string{"x", "CI-agent", "ci_agent", "ci agent", "-ci", "ci-", "ci--agent", strings.Repeat("a", 40)} {
		if !stringAttributeErrors(t, s, "name", bad) {
			t.Errorf("name %q: expected a validation error", bad)
		}
	}
	for _, good := range []string{"ci", "ci-agent", "run-triage-2", strings.Repeat("a", 39)} {
		if stringAttributeErrors(t, s, "name", good) {
			t.Errorf("name %q: unexpectedly rejected", good)
		}
	}
}
