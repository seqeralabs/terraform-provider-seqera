package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ComputeenvStateUpgraderV0 migrates seqera_compute_env state from schema
// version 0 (e.g. provider v0.25.x) directly to the current schema. The plugin
// framework does not chain upgraders, so this applies every migration between v0
// and the current schema:
//
//   - v0 -> v1: rename the misspelled `nvnme_storage_enabled` flag to
//     `nvme_storage_enabled` (renameNvmeStorageFlag).
//   - v1 -> current: derive the Azure Batch delete_jobs_on_completion boolean
//     (applyComputeEnvV2Migrations).
//
// Attribute removals are not enumerated — upgradeToCurrentSchema drops every
// attribute absent from the current schema.
func ComputeenvStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_compute_env", req, resp, func(rawState map[string]interface{}) {
		// Rename the misspelled nvme flag before re-decoding; otherwise the
		// misspelled key (absent from the current schema) would be dropped
		// instead of carried to the new name.
		renameNvmeStorageFlag(rawState)

		if computeEnv, ok := rawState["compute_env"].(map[string]interface{}); ok {
			applyComputeEnvV2Migrations(computeEnv)
		}
	})
}
