package stringvalidators

import (
	"context"
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

// allKeyNames is the single source of truth for credential key block names in tests.
var allKeyNames = []string{
	"aws", "azure", "azure_cloud", "azure_entra", "azurerepos",
	"bitbucket", "codecommit", "container_reg", "gitea", "github",
	"gitlab", "google", "k8s", "local", "s3", "ssh", "seqeracompute", "tw_agent",
}

// keyBlockTFType is the tftypes representation of a single key block (object with "id" string).
var keyBlockTFType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String}}

// keysObjectType returns the tftypes for the keys object containing all key blocks.
func keysObjectType() tftypes.Object {
	attrs := make(map[string]tftypes.Type, len(allKeyNames))
	for _, name := range allKeyNames {
		attrs[name] = keyBlockTFType
	}
	return tftypes.Object{AttributeTypes: attrs}
}

// credentialObjectType returns the top-level tftypes containing provider_type and keys.
func credentialObjectType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"provider_type": tftypes.String,
			"keys":          keysObjectType(),
		},
	}
}

// buildKeysValue builds a tftypes.Value for the keys object.
// The specified setKeys are populated; all others are null.
func buildKeysValue(setKeys ...string) tftypes.Value {
	kt := keysObjectType()
	set := make(map[string]bool, len(setKeys))
	for _, k := range setKeys {
		set[k] = true
	}
	vals := make(map[string]tftypes.Value, len(allKeyNames))
	for _, name := range allKeyNames {
		if set[name] {
			vals[name] = tftypes.NewValue(keyBlockTFType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "test-id"),
			})
		} else {
			vals[name] = tftypes.NewValue(keyBlockTFType, nil)
		}
	}
	return tftypes.NewValue(kt, vals)
}

// keysNestedAttribute builds the resource schema attribute for the keys object.
func keysNestedAttribute() resourceschema.SingleNestedAttribute {
	block := resourceschema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{Optional: true},
		},
	}
	attrs := make(map[string]resourceschema.Attribute, len(allKeyNames))
	for _, name := range allKeyNames {
		attrs[name] = block
	}
	return resourceschema.SingleNestedAttribute{
		Required:   true,
		Attributes: attrs,
	}
}

// makeRequest constructs a validator.StringRequest with the given provider_type and keys value.
func makeRequest(providerType string, keysVal tftypes.Value) validator.StringRequest {
	ct := credentialObjectType()
	rawVal := tftypes.NewValue(ct, map[string]tftypes.Value{
		"provider_type": tftypes.NewValue(tftypes.String, providerType),
		"keys":          keysVal,
	})
	config := tfsdk.Config{
		Schema: resourceschema.Schema{
			Attributes: map[string]resourceschema.Attribute{
				"provider_type": resourceschema.StringAttribute{Required: true},
				"keys":          keysNestedAttribute(),
			},
		},
		Raw: rawVal,
	}
	return validator.StringRequest{
		Path:        path.Root("provider_type"),
		ConfigValue: types.StringValue(providerType),
		Config:      config,
	}
}

// runStringValidator runs any string validator and returns diagnostics.
func runStringValidator(v validator.String, req validator.StringRequest) diag.Diagnostics {
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	return resp.Diagnostics
}

func runValidator(req validator.StringRequest) diag.Diagnostics {
	return runStringValidator(CredentialsConfigValidator(), req)
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------

func TestCredentialsConfigValidator_ValidSingleKeyProviders(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"aws":           "aws",
		"azure":         "azure",
		"google":        "google",
		"github":        "github",
		"gitlab":        "gitlab",
		"bitbucket":     "bitbucket",
		"ssh":           "ssh",
		"k8s":           "k8s",
		"container-reg": "container_reg",
		"tw-agent":      "tw_agent",
		"codecommit":    "codecommit",
		"gitea":         "gitea",
		"azurerepos":    "azurerepos",
		"seqeracompute": "seqeracompute",
		"azure_entra":   "azure_entra",
	}

	for provider, keyField := range cases {
		t.Run(provider, func(t *testing.T) {
			t.Parallel()
			req := makeRequest(provider, buildKeysValue(keyField))
			diags := runValidator(req)
			assert.False(t, diags.HasError(), "expected no errors for provider_type=%q with keys.%s, got: %s", provider, keyField, diags.Errors())
		})
	}
}

func TestCredentialsConfigValidator_AzureWithAzureCloudKeys(t *testing.T) {
	t.Parallel()
	req := makeRequest("azure", buildKeysValue("azure_cloud"))
	diags := runValidator(req)
	assert.False(t, diags.HasError(), "provider_type=azure should accept keys.azure_cloud, got: %s", diags.Errors())
}

func TestCredentialsConfigValidator_AzureWithAzureKeys(t *testing.T) {
	t.Parallel()
	req := makeRequest("azure", buildKeysValue("azure"))
	diags := runValidator(req)
	assert.False(t, diags.HasError(), "provider_type=azure should accept keys.azure, got: %s", diags.Errors())
}

func TestCredentialsConfigValidator_InvalidProviderType(t *testing.T) {
	t.Parallel()
	req := makeRequest("invalid_provider", buildKeysValue("aws"))
	diags := runValidator(req)
	assert.True(t, diags.HasError(), "expected error for invalid provider_type")
	assert.Contains(t, diags.Errors()[0].Summary(), "Invalid Provider Type")
}

func TestCredentialsConfigValidator_MismatchedKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		provider string
		keyField string
	}{
		{"aws with google keys", "aws", "google"},
		{"google with aws keys", "google", "aws"},
		{"azure with github keys", "azure", "github"},
		{"azure_entra with azure keys", "azure_entra", "azure"},
		{"ssh with k8s keys", "ssh", "k8s"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := makeRequest(tc.provider, buildKeysValue(tc.keyField))
			diags := runValidator(req)
			assert.True(t, diags.HasError(), "expected mismatch error for provider_type=%q with keys.%s", tc.provider, tc.keyField)
			assert.Contains(t, diags.Errors()[0].Summary(), "Provider Keys Mismatch")
		})
	}
}

func TestCredentialsConfigValidator_MissingKeys(t *testing.T) {
	t.Parallel()
	req := makeRequest("aws", buildKeysValue()) // no keys set
	diags := runValidator(req)
	assert.True(t, diags.HasError(), "expected error when no keys are set")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Provider Keys")
}

func TestCredentialsConfigValidator_MultipleKeysSet(t *testing.T) {
	t.Parallel()
	req := makeRequest("aws", buildKeysValue("aws", "google"))
	diags := runValidator(req)
	assert.True(t, diags.HasError(), "expected error when multiple keys are set")
	assert.Contains(t, diags.Errors()[0].Summary(), "Multiple Provider Keys")
}

func TestCredentialsConfigValidator_NullProviderType(t *testing.T) {
	t.Parallel()
	req := validator.StringRequest{
		Path:        path.Root("provider_type"),
		ConfigValue: types.StringNull(),
	}
	diags := runValidator(req)
	assert.False(t, diags.HasError(), "null provider_type should skip validation")
}

func TestCredentialsConfigValidator_UnknownProviderType(t *testing.T) {
	t.Parallel()
	req := validator.StringRequest{
		Path:        path.Root("provider_type"),
		ConfigValue: types.StringUnknown(),
	}
	diags := runValidator(req)
	assert.False(t, diags.HasError(), "unknown provider_type should skip validation")
}

func TestCredentialsConfigValidator_AzureBothAzureAndCloudKeysRejected(t *testing.T) {
	t.Parallel()
	// Setting both keys.azure and keys.azure_cloud simultaneously should fail.
	req := makeRequest("azure", buildKeysValue("azure", "azure_cloud"))
	diags := runValidator(req)
	assert.True(t, diags.HasError(), "provider_type=azure with both keys.azure and keys.azure_cloud should fail")
	assert.Contains(t, diags.Errors()[0].Summary(), "Multiple Provider Keys")
}

func TestCredentialsConfigValidator_AzureRejectsEntraKeys(t *testing.T) {
	t.Parallel()
	req := makeRequest("azure", buildKeysValue("azure_entra"))
	diags := runValidator(req)
	assert.True(t, diags.HasError(), "provider_type=azure should reject keys.azure_entra")
	assert.Contains(t, diags.Errors()[0].Summary(), "Provider Keys Mismatch")
}

func TestCredentialsConfigValidator_AzureEntraRejectsAzureCloudKeys(t *testing.T) {
	t.Parallel()
	req := makeRequest("azure_entra", buildKeysValue("azure_cloud"))
	diags := runValidator(req)
	assert.True(t, diags.HasError(), "provider_type=azure_entra should reject keys.azure_cloud")
	assert.Contains(t, diags.Errors()[0].Summary(), "Provider Keys Mismatch")
}

// -----------------------------------------------------------------------
// Unit tests for helper functions
// -----------------------------------------------------------------------

func TestSchemaNameToKeysFieldName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"AgentSecurityKeys":         "tw_agent",
		"AwsSecurityKeys":           "aws",
		"AzureCloudSecurityKeys":    "azure_cloud",
		"AzureEntraKeys":            "azure_entra",
		"AzureReposSecurityKeys":    "azurerepos",
		"AzureSecurityKeys":         "azure",
		"BitBucketSecurityKeys":     "bitbucket",
		"CodeCommitSecurityKeys":    "codecommit",
		"ContainerRegistryKeys":     "container_reg",
		"GiteaSecurityKeys":         "gitea",
		"GitHubSecurityKeys":        "github",
		"GitLabSecurityKeys":        "gitlab",
		"GoogleSecurityKeys":        "google",
		"K8sSecurityKeys":           "k8s",
		"SeqeraComputeSecurityKeys": "seqeracompute",
		"SSHSecurityKeys":           "ssh",
	}

	for schemaName, expected := range cases {
		t.Run(schemaName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, expected, schemaNameToKeysFieldName(schemaName))
		})
	}
}

func TestSchemaNameToKeysFieldName_Fallback(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "unknowntype", schemaNameToKeysFieldName("UnknownType"))
}

func TestMapKeys(t *testing.T) {
	t.Parallel()
	m := map[string][]string{
		"aws":   {"AwsSecurityKeys"},
		"azure": {"AzureSecurityKeys", "AzureCloudSecurityKeys"},
	}
	keys := mapKeys(m)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "aws")
	assert.Contains(t, keys, "azure")
}
