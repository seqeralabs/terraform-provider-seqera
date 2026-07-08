package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ManagedcomputeceStateUpgraderV0 upgrades seqera_managed_compute_ce state to the current schema by re-decoding prior state
// against it, dropping any attribute the schema no longer defines. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func ManagedcomputeceStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_managed_compute_ce", req, resp, nil)
}
