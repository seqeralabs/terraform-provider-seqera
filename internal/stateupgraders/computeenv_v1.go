package stateupgraders

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ComputeenvStateUpgraderV1 migrates the state from version 1 to version 2.
//
// Azure Batch breaking change: in v0.30.x, `config.azure_batch.delete_jobs_on_completion`
// was a settable string field (e.g. "on_success", "on_failure"). In v0.40.0 the field
// is read-only and is replaced by three new boolean fields:
//   - delete_jobs_on_completion_enabled
//   - delete_pools_on_completion
//   - delete_tasks_on_completion
//
// Without this upgrader, users carrying old state would see a spurious "null -> true"
// diff on `delete_jobs_on_completion_enabled` after they update their config, which
// would force a resource replacement. We translate any non-empty old string value
// into delete_jobs_on_completion_enabled = true so the plan stays clean.
func ComputeenvStateUpgraderV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var rawState map[string]interface{}
	err := json.Unmarshal(req.RawState.JSON, &rawState)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Unmarshal Prior State",
			err.Error(),
		)
		return
	}

	// Navigate to config.azure_batch and migrate delete_jobs_on_completion.
	if config, ok := rawState["config"].(map[string]interface{}); ok {
		if azureBatch, ok := config["azure_batch"].(map[string]interface{}); ok {
			if oldValue, exists := azureBatch["delete_jobs_on_completion"]; exists {
				if s, ok := oldValue.(string); ok && s != "" {
					// Only set the new flag if the user previously asked for cleanup.
					// Don't overwrite an explicit false the user may have set.
					if _, alreadySet := azureBatch["delete_jobs_on_completion_enabled"]; !alreadySet {
						azureBatch["delete_jobs_on_completion_enabled"] = true
					}
				}
			}
		}
	}

	upgradedStateJSON, err := json.Marshal(rawState)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Marshal Upgraded State",
			err.Error(),
		)
		return
	}

	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: upgradedStateJSON,
	}
}
