package stateupgraders

import "github.com/hashicorp/terraform-plugin-framework/resource"
import "context"

func AzurecloudceStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	// No-op
}
