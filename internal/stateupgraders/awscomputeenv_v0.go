package stateupgraders

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// AwscomputeenvStateUpgraderV0 migrates the state from version 0 to version 1
// This handles the field rename from nvnme_storage_enabled to nvme_storage_enabled
func AwscomputeenvStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	// Unmarshal the raw state (JSON format from Terraform state file)
	var rawState map[string]interface{}
	err := json.Unmarshal(req.RawState.JSON, &rawState)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Unmarshal Prior State",
			err.Error(),
		)
		return
	}

	// Navigate to the config object and perform the migration
	if config, ok := rawState["config"].(map[string]interface{}); ok {
		if oldValue, exists := config["nvnme_storage_enabled"]; exists {
			// Copy the value to the new field name
			config["nvme_storage_enabled"] = oldValue
			// Remove the old field name
			delete(config, "nvnme_storage_enabled")
		}
	}

	// Marshal the updated state back to JSON
	upgradedStateJSON, err := json.Marshal(rawState)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Marshal Upgraded State",
			err.Error(),
		)
		return
	}

	// Set the upgraded state as raw JSON
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: upgradedStateJSON,
	}
}
