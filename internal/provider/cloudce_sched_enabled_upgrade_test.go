package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/seqeralabs/terraform-provider-seqera/internal/stateupgraders"
)

type upgraderFunc = func(context.Context, resource.UpgradeStateRequest, *resource.UpgradeStateResponse)

// TestCloudCEUpgrade_RenamesSchedEnabled covers the rename of `sched_enabled` to
// `intelligent_compute_enabled` on seqera_azure_cloud_ce and seqera_gcp_cloud_ce.
// Every registered upgrader must carry the value across, since the framework
// does not chain upgraders.
func TestCloudCEUpgrade_RenamesSchedEnabled(t *testing.T) {
	ctx := context.Background()

	resources := []struct {
		name        string
		newResource func() resource.Resource
		upgraders   map[string]upgraderFunc
	}{
		{"azure_cloud_ce", NewAzureCloudCEResource, map[string]upgraderFunc{
			"v0": stateupgraders.AzurecloudceStateUpgraderV0,
			"v1": stateupgraders.AzurecloudceStateUpgraderV1,
		}},
		{"gcp_cloud_ce", NewGCPCloudCEResource, map[string]upgraderFunc{
			"v0": stateupgraders.GcpcloudceStateUpgraderV0,
			"v1": stateupgraders.GcpcloudceStateUpgraderV1,
		}},
	}

	for _, r := range resources {
		for version, upgrade := range r.upgraders {
			t.Run(r.name+"/"+version, func(t *testing.T) {
				state := runUpgraderAgainstSchema(t, r.newResource, upgrade, map[string]interface{}{
					"name":   "my-ce",
					"config": map[string]interface{}{"sched_enabled": true},
				})

				var enabled types.Bool
				diags := state.GetAttribute(ctx, path.Root("config").AtName("intelligent_compute_enabled"), &enabled)
				if diags.HasError() {
					t.Fatalf("reading intelligent_compute_enabled: %v", diags)
				}
				if enabled.IsNull() || !enabled.ValueBool() {
					t.Errorf("expected intelligent_compute_enabled=true, got %v", enabled)
				}
			})
		}
	}
}

// TestComputeenvUpgrade_RenamesSchedEnabled covers the same rename inside the
// azure_cloud and google_cloud blocks of seqera_compute_env, for every
// registered upgrader.
func TestComputeenvUpgrade_RenamesSchedEnabled(t *testing.T) {
	ctx := context.Background()

	for _, platform := range []string{"azure_cloud", "google_cloud"} {
		for version, upgrade := range computeEnvUpgraders {
			t.Run(platform+"/"+version, func(t *testing.T) {
				state := runUpgraderAgainstSchema(t, NewComputeEnvResource, upgrade, map[string]interface{}{
					"compute_env": map[string]interface{}{
						"name": "my-ce",
						"config": map[string]interface{}{
							platform: map[string]interface{}{"sched_enabled": true},
						},
					},
				})

				var enabled types.Bool
				diags := state.GetAttribute(ctx, path.Root("compute_env").AtName("config").AtName(platform).AtName("intelligent_compute_enabled"), &enabled)
				if diags.HasError() {
					t.Fatalf("reading intelligent_compute_enabled: %v", diags)
				}
				if enabled.IsNull() || !enabled.ValueBool() {
					t.Errorf("expected intelligent_compute_enabled=true, got %v", enabled)
				}
			})
		}
	}
}
