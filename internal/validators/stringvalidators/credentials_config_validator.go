package stringvalidators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringCredentialsConfigValidatorValidator{}

type StringCredentialsConfigValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringCredentialsConfigValidatorValidator) Description(_ context.Context) string {
	return "validates that the provider_type value matches the corresponding keys schema type"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringCredentialsConfigValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringCredentialsConfigValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if value is null or unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	providerType := req.ConfigValue.ValueString()

	// Get the parent object to access the keys field
	parentPath := req.Path.ParentPath()
	var parentObj types.Object

	// Get the parent object from the configuration
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, parentPath, &parentObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract the keys field from the parent object
	keysPath := parentPath.AtName("keys")
	var keysObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, keysPath, &keysObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if keys is null or unknown
	if keysObj.IsNull() || keysObj.IsUnknown() {
		return
	}

	// Get the expected keys field name by converting schema name to field name
	// Provider to schema mapping from the discriminator in the OpenAPI spec:
	providerKeysMap := map[string]string{
		"aws":           "AwsSecurityKeys",
		"azure":         "AzureSecurityKeys",
		"google":        "GoogleSecurityKeys",
		"github":        "GitHubSecurityKeys",
		"gitlab":        "GitLabSecurityKeys",
		"bitbucket":     "BitBucketSecurityKeys",
		"ssh":           "SSHSecurityKeys",
		"k8s":           "K8sSecurityKeys",
		"container-reg": "ContainerRegistryKeys",
		"tw-agent":      "AgentSecurityKeys",
		"codecommit":    "CodeCommitSecurityKeys",
		"gitea":         "GiteaSecurityKeys",
		"azurerepos":    "AzureReposSecurityKeys",
		"seqeracompute": "SeqeraComputeSecurityKeys",
		"azure_entra":   "AzureEntraKeys",
	}

	// Get the expected keys schema name for this provider
	expectedSchemaName, exists := providerKeysMap[providerType]
	if !exists {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Provider Type",
			fmt.Sprintf("Provider type '%s' is not supported. Valid provider types are: %v", providerType, getValidProviders(providerKeysMap)),
		)
		return
	}

	// Convert schema name to field name (e.g., "AwsSecurityKeys" -> "aws")
	expectedFieldName := schemaNameToKeysFieldName(expectedSchemaName)

	// Check if the keys object has exactly one keys field set and it matches the provider
	keysAttrs := keysObj.Attributes()
	matchCount := 0
	var setKeysNames []string

	for keysName, attr := range keysAttrs {
		if !attr.IsNull() && !attr.IsUnknown() {
			setKeysNames = append(setKeysNames, keysName)
			if keysName == expectedFieldName {
				matchCount++
			}
		}
	}

	// Check for mismatched keys (keys set but doesn't match provider)
	if len(setKeysNames) > 0 && matchCount == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Provider Keys Mismatch",
			fmt.Sprintf("Provider type '%s' cannot be used with keys type '%s'. Expected keys type: '%s'", providerType, setKeysNames[0], expectedFieldName),
		)
		return
	}

	// Check for multiple keys set
	if len(setKeysNames) > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Multiple Provider Keys",
			fmt.Sprintf("Only one keys type can be set, but found multiple: %v", setKeysNames),
		)
		return
	}

	// Check for missing keys
	if len(setKeysNames) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Provider Keys",
			fmt.Sprintf("Provider type '%s' requires keys type '%s' to be set", providerType, expectedFieldName),
		)
	}
}

// Helper function to convert schema name to keys field name
// Based on the actual generated field names in SecurityKeysOutput
func schemaNameToKeysFieldName(schemaName string) string {
	// Map schema names directly to the field names used in SecurityKeysOutput
	switch schemaName {
	case "AgentSecurityKeys":
		return "tw_agent"
	case "AwsSecurityKeys":
		return "aws"
	case "AzureEntraKeys":
		return "azure_entra"
	case "AzureReposSecurityKeys":
		return "azurerepos"
	case "AzureSecurityKeys":
		return "azure"
	case "BitBucketSecurityKeys":
		return "bitbucket"
	case "CodeCommitSecurityKeys":
		return "codecommit"
	case "ContainerRegistryKeys":
		return "container_reg"
	case "GiteaSecurityKeys":
		return "gitea"
	case "GitHubSecurityKeys":
		return "github"
	case "GitLabSecurityKeys":
		return "gitlab"
	case "GoogleSecurityKeys":
		return "google"
	case "K8sSecurityKeys":
		return "k8s"
	case "SeqeraComputeSecurityKeys":
		return "seqeracompute"
	case "SSHSecurityKeys":
		return "ssh"
	default:
		// Fallback to lowercase
		return strings.ToLower(schemaName)
	}
}

// Helper function to get valid providers
func getValidProviders(providerMap map[string]string) []string {
	providers := make([]string, 0, len(providerMap))
	for provider := range providerMap {
		providers = append(providers, provider)
	}
	return providers
}

func CredentialsConfigValidator() validator.String {
	return StringCredentialsConfigValidatorValidator{}
}
