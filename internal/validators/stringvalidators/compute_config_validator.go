package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringComputeConfigValidatorValidator{}

type StringComputeConfigValidatorValidator struct{}

func (v StringComputeConfigValidatorValidator) Description(_ context.Context) string {
	return "validates that the platform value matches the corresponding config schema type"
}

func (v StringComputeConfigValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringComputeConfigValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	platform := req.ConfigValue.ValueString()

	// Get the parent object to access the config field
	parentPath := req.Path.ParentPath()
	var parentObj types.Object

	// Get the parent object from the configuration
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, parentPath, &parentObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract the config field from the parent object
	configPath := parentPath.AtName("config")
	var configObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, configPath, &configObj)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if config is null or unknown
	if configObj.IsNull() || configObj.IsUnknown() {
		return
	}

	// Define the platform to config schema mapping
	// TODO: We can probably import this from the OpenAPI spec instead of hardcoding it
	platformConfigMap := map[string][]string{
		"aws-batch":              {"aws_batch_configuration"},
		"aws-cloud":              {"aws_cloud_configuration"},
		"azure-batch":            {"azure_batch_configuration"},
		"azure-cloud":            {"azure_cloud_configuration"},
		"google-lifesciences":    {"google_life_sciences_configuration_retired"},
		"google-batch":           {"google_batch_service_configuration"},
		"google-cloud":           {"google_cloud_configuration"},
		"seqeracompute-platform": {"seqera_compute_configuration"},
		"k8s-platform":           {"kubernetes_compute_configuration"},
		"eks-platform":           {"amazon_eks_cluster_configuration"},
		"gke-platform":           {"google_gke_cluster_configuration"},
		"slurm-platform":         {"slurm_configuration"},
		"lsf-platform":           {"ibmlsf_configuration"},
		"altair-platform":        {"altair_pbs_configuration"},
		"moab-platform":          {"moab_configuration"},
		"uge-platform":           {"univa_grid_engine_configuration"},
	}

	// Get the expected config field names for this platform
	expectedConfigs, exists := platformConfigMap[platform]
	if !exists {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Platform",
			fmt.Sprintf("Platform '%s' is not supported. Valid platforms are: %v", platform, getValidPlatforms(platformConfigMap)),
		)
		return
	}

	// Check if the config object has exactly one of the expected config fields set
	configAttrs := configObj.Attributes()
	matchCount := 0

	for _, expectedConfig := range expectedConfigs {
		if attr, ok := configAttrs[expectedConfig]; ok && !attr.IsNull() && !attr.IsUnknown() {
			matchCount++
		}
	}

	// Also check for any config fields that don't match the platform
	for configName, attr := range configAttrs {
		if !attr.IsNull() && !attr.IsUnknown() {
			// Check if this config belongs to the current platform
			belongsToCurrentPlatform := false
			for _, expectedConfig := range expectedConfigs {
				if configName == expectedConfig {
					belongsToCurrentPlatform = true
					break
				}
			}
			if !belongsToCurrentPlatform {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Platform Configuration Mismatch",
					fmt.Sprintf("Platform '%s' cannot be used with config type '%s'. Expected config types for platform '%s': %v", platform, configName, platform, expectedConfigs),
				)
				return
			}
		}
	}

	if matchCount == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Platform Configuration",
			fmt.Sprintf("Platform '%s' requires one of the following config types to be set: %v", platform, expectedConfigs),
		)
	} else if matchCount > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Multiple Platform Configurations",
			fmt.Sprintf("Platform '%s' can only have one config type set, but found multiple configurations", platform),
		)
	}
}

// Helper function to get valid platforms
func getValidPlatforms(platformMap map[string][]string) []string {
	platforms := make([]string, 0, len(platformMap))
	for platform := range platformMap {
		platforms = append(platforms, platform)
	}
	return platforms
}

func ComputeConfigValidator() validator.String {
	return StringComputeConfigValidatorValidator{}
}
