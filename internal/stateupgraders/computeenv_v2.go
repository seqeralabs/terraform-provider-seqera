package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ComputeenvStateUpgraderV2 migrates seqera_compute_env state from schema
// version 2 to the current schema: `sched_enabled` in the azure_cloud and
// google_cloud config blocks was renamed to `intelligent_compute_enabled` to
// match aws_cloud (applyComputeEnvV3Migrations).
func ComputeenvStateUpgraderV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_compute_env", req, resp, func(rawState map[string]interface{}) {
		if computeEnv, ok := rawState["compute_env"].(map[string]interface{}); ok {
			applyComputeEnvV3Migrations(computeEnv)
		}
	})
}

// applyComputeEnvV3Migrations renames `sched_enabled` to
// `intelligent_compute_enabled` in the azure_cloud and google_cloud config
// blocks of a prior `compute_env` object. Shared by the v0, v1 and v2 upgraders
// (the framework does not chain upgraders).
func applyComputeEnvV3Migrations(computeEnv map[string]interface{}) {
	config, ok := computeEnv["config"].(map[string]interface{})
	if !ok {
		return
	}
	for _, platform := range []string{"azure_cloud", "google_cloud"} {
		if platformConfig, ok := config[platform].(map[string]interface{}); ok {
			renameSchedEnabledFlag(platformConfig)
		}
	}
}
