package stringvalidators

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func googleCredObjectType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"data":                       tftypes.String,
			"workload_identity_provider": tftypes.String,
			"service_account_email":      tftypes.String,
		},
	}
}

func makeGoogleRequest(fieldName string, values map[string]*string) validator.StringRequest {
	ct := googleCredObjectType()

	tfvals := make(map[string]tftypes.Value)
	for _, name := range []string{"data", "workload_identity_provider", "service_account_email"} {
		if v, ok := values[name]; ok && v != nil {
			tfvals[name] = tftypes.NewValue(tftypes.String, *v)
		} else {
			tfvals[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}

	rawVal := tftypes.NewValue(ct, tfvals)

	s := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"data":                       resourceschema.StringAttribute{Optional: true},
			"workload_identity_provider": resourceschema.StringAttribute{Optional: true},
			"service_account_email":      resourceschema.StringAttribute{Optional: true},
		},
	}

	config := tfsdk.Config{
		Schema: s,
		Raw:    rawVal,
	}

	var configValue types.String
	if v, ok := values[fieldName]; ok && v != nil {
		configValue = types.StringValue(*v)
	} else {
		configValue = types.StringNull()
	}

	return validator.StringRequest{
		Path:        path.Root(fieldName),
		ConfigValue: configValue,
		Config:      config,
	}
}

func runGoogleValidator(req validator.StringRequest) diag.Diagnostics {
	return runStringValidator(GoogleCredentialKeysValidator(), req)
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------

func TestGoogleCredentialKeys_DataOnly(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"data":                       strPtr(`{"type":"service_account"}`),
		"workload_identity_provider": nil,
		"service_account_email":      nil,
	}
	diags := runGoogleValidator(makeGoogleRequest("data", values))
	assert.False(t, diags.HasError(), "data alone should be valid, got: %s", diags.Errors())
}

func TestGoogleCredentialKeys_WIFPair(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"data":                       nil,
		"workload_identity_provider": strPtr("projects/123456789012/locations/global/workloadIdentityPools/p/providers/pr"),
		"service_account_email":      strPtr("seqera@p.iam.gserviceaccount.com"),
	}
	for _, field := range []string{"workload_identity_provider", "service_account_email"} {
		t.Run("validate_"+field, func(t *testing.T) {
			t.Parallel()
			diags := runGoogleValidator(makeGoogleRequest(field, values))
			assert.False(t, diags.HasError(), "WIF pair should be valid from %s, got: %s", field, diags.Errors())
		})
	}
}

func TestGoogleCredentialKeys_DataWithWIF(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"data":                       strPtr(`{"type":"service_account"}`),
		"workload_identity_provider": strPtr("projects/123456789012/locations/global/workloadIdentityPools/p/providers/pr"),
		"service_account_email":      strPtr("seqera@p.iam.gserviceaccount.com"),
	}
	diags := runGoogleValidator(makeGoogleRequest("data", values))
	assert.True(t, diags.HasError(), "data + WIF should fail as conflicting")
	assert.Contains(t, diags.Errors()[0].Summary(), "Conflicting Authentication Methods")
}

func TestGoogleCredentialKeys_WIFProviderWithoutEmail(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"data":                       nil,
		"workload_identity_provider": strPtr("projects/123456789012/locations/global/workloadIdentityPools/p/providers/pr"),
		"service_account_email":      nil,
	}
	diags := runGoogleValidator(makeGoogleRequest("workload_identity_provider", values))
	assert.True(t, diags.HasError(), "workload_identity_provider without service_account_email should fail")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Attribute")
}

func TestGoogleCredentialKeys_EmailWithoutWIFProvider(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"data":                       nil,
		"workload_identity_provider": nil,
		"service_account_email":      strPtr("seqera@p.iam.gserviceaccount.com"),
	}
	diags := runGoogleValidator(makeGoogleRequest("service_account_email", values))
	assert.True(t, diags.HasError(), "service_account_email without workload_identity_provider should fail")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Attribute")
}

func TestGoogleCredentialKeys_NothingProvided(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"data":                       nil,
		"workload_identity_provider": nil,
		"service_account_email":      nil,
	}
	diags := runGoogleValidator(makeGoogleRequest("data", values))
	assert.True(t, diags.HasError(), "no credentials at all should fail")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Configuration")
}

func TestGoogleCredentialKeys_UnknownValuesSkipsValidation(t *testing.T) {
	t.Parallel()
	ct := googleCredObjectType()
	rawVal := tftypes.NewValue(ct, map[string]tftypes.Value{
		"data":                       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"workload_identity_provider": tftypes.NewValue(tftypes.String, nil),
		"service_account_email":      tftypes.NewValue(tftypes.String, nil),
	})

	s := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"data":                       resourceschema.StringAttribute{Optional: true},
			"workload_identity_provider": resourceschema.StringAttribute{Optional: true},
			"service_account_email":      resourceschema.StringAttribute{Optional: true},
		},
	}

	req := validator.StringRequest{
		Path:        path.Root("data"),
		ConfigValue: types.StringUnknown(),
		Config:      tfsdk.Config{Schema: s, Raw: rawVal},
	}

	diags := runGoogleValidator(req)
	assert.False(t, diags.HasError(), "unknown values should skip validation, got: %s", diags.Errors())
}

func TestGoogleCredentialKeys_EmptyStringsCountAsUnset(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"data":                       strPtr(""),
		"workload_identity_provider": strPtr(""),
		"service_account_email":      strPtr(""),
	}
	diags := runGoogleValidator(makeGoogleRequest("data", values))
	assert.True(t, diags.HasError(), "empty strings should be treated as unset")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Configuration")
}
