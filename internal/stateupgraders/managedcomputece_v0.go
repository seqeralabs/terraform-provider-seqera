package stateupgraders

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"context"
)

func ManagedcomputeceStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	// No-op
}
