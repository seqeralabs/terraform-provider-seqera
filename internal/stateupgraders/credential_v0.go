package stateupgraders

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// CredentialStateUpgraderV0 migrates the state from version 0 to version 1
// This handles the field rename from credentials_id to id
func CredentialStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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

	// Rename credentials_id to id at the root level
	if oldValue, exists := rawState["credentials_id"]; exists {
		// Copy the value to the new field name
		rawState["id"] = oldValue
		// Remove the old field name
		delete(rawState, "credentials_id")
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
