package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// GitlabcredentialStateUpgraderV0 migrates the state from version 0 to version 1
// This is a no-op upgrade: the credential id field mapping changed at the SDK
// layer (JSON tag), but the Terraform attribute name remains credentials_id.
func GitlabcredentialStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: req.RawState.JSON,
	}
}
