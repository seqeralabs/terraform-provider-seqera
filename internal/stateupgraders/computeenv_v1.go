package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ComputeenvStateUpgraderV1 migrates seqera_compute_env state from schema
// version 1 (e.g. provider v0.30.x) to the current schema (issue #228).
//
// Two classes of change occur between v1 and the current schema:
//
//   - Attribute removals — the top-level `compute_env.deleted` bool, and various
//     per-platform config fields (e.g. `config.aws_cloud.enable_fusion`,
//     `enable_wave`). These are NOT handled explicitly; upgradeToCurrentSchema
//     re-decodes against the current schema and drops every removed attribute at
//     any nesting depth.
//
//   - Value transforms — the Azure Batch `delete_jobs_on_completion` string was
//     superseded by the boolean `delete_jobs_on_completion_enabled`, which needs
//     an explicit derivation (applyComputeEnvV2Migrations).
func ComputeenvStateUpgraderV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_compute_env", req, resp, func(rawState map[string]interface{}) {
		if computeEnv, ok := rawState["compute_env"].(map[string]interface{}); ok {
			applyComputeEnvV2Migrations(computeEnv)
		}
	})
}

// applyComputeEnvV2Migrations performs the *value* transforms needed to carry a
// prior `compute_env` object forward to the current schema. Attribute *removals*
// are intentionally not handled here — upgradeToCurrentSchema drops every
// attribute absent from the current schema automatically. This function only
// handles changes that derive a new value from an old one, and is shared by the
// v0 and v1 upgraders (the framework does not chain upgraders, so each migrates
// its version directly to the current schema).
func applyComputeEnvV2Migrations(computeEnv map[string]interface{}) {
	config, ok := computeEnv["config"].(map[string]interface{})
	if !ok {
		return
	}
	azureBatch, ok := config["azure_batch"].(map[string]interface{})
	if !ok {
		return
	}

	// Azure Batch: the legacy settable string `delete_jobs_on_completion`
	// ("on_success", "always", "never") was replaced by the boolean
	// `delete_jobs_on_completion_enabled`. Derive the boolean from the old string
	// so the plan stays clean, but never overwrite a value the user already set.
	oldValue, exists := azureBatch["delete_jobs_on_completion"]
	if !exists {
		return
	}
	if _, alreadySet := azureBatch["delete_jobs_on_completion_enabled"]; alreadySet {
		return
	}
	if s, ok := oldValue.(string); ok {
		switch s {
		case "always", "on_success":
			azureBatch["delete_jobs_on_completion_enabled"] = true
		case "never":
			azureBatch["delete_jobs_on_completion_enabled"] = false
		}
	}
}
