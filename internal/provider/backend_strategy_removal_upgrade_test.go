package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/seqeralabs/terraform-provider-seqera/internal/stateupgraders"
)

// intelligentComputeState is a config fragment as stored by provider versions
// that still exposed intelligent_compute_config.backend_strategy.
func intelligentComputeState() map[string]interface{} {
	return map[string]interface{}{
		"intelligent_compute_config": map[string]interface{}{
			"backend_strategy":   "ECS",
			"provisioning_model": "spotFirst",
		},
	}
}

// TestUpgrade_DropsRemovedBackendStrategy checks that every registered state
// upgrader of the Intelligent Compute resources drops the removed
// backend_strategy attribute, since runUpgraderAgainstSchema decodes strictly
// against the current schema. (State at the current schema version is
// decoded by the framework with IgnoreUndefinedAttributes.)
func TestUpgrade_DropsRemovedBackendStrategy(t *testing.T) {
	cloudCEs := []struct {
		name        string
		newResource func() resource.Resource
		upgraders   map[string]upgraderFunc
	}{
		{"aws_cloud_ce", NewAwsCloudCEResource, map[string]upgraderFunc{
			"v0": stateupgraders.AwscloudceStateUpgraderV0,
		}},
		{"azure_cloud_ce", NewAzureCloudCEResource, map[string]upgraderFunc{
			"v0": stateupgraders.AzurecloudceStateUpgraderV0,
			"v1": stateupgraders.AzurecloudceStateUpgraderV1,
		}},
		{"gcp_cloud_ce", NewGCPCloudCEResource, map[string]upgraderFunc{
			"v0": stateupgraders.GcpcloudceStateUpgraderV0,
			"v1": stateupgraders.GcpcloudceStateUpgraderV1,
		}},
	}
	for _, r := range cloudCEs {
		for version, upgrade := range r.upgraders {
			t.Run(r.name+"/"+version, func(t *testing.T) {
				runUpgraderAgainstSchema(t, r.newResource, upgrade, map[string]interface{}{
					"name":   "my-ce",
					"config": intelligentComputeState(),
				})
			})
		}
	}

	for _, platform := range []string{"aws_cloud", "azure_cloud", "google_cloud", "local_platform"} {
		for version, upgrade := range computeEnvUpgraders {
			t.Run("compute_env/"+platform+"/"+version, func(t *testing.T) {
				runUpgraderAgainstSchema(t, NewComputeEnvResource, upgrade, map[string]interface{}{
					"compute_env": map[string]interface{}{
						"name":   "my-ce",
						"config": map[string]interface{}{platform: intelligentComputeState()},
					},
				})
			})
		}
	}
}
