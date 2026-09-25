package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// GcpcloudceStateUpgraderV0 migrates seqera_gcp_cloud_ce state from schema version 0 directly to
// the current schema. The framework does not chain upgraders, so this applies
// the v1 -> v2 rename of `config.sched_enabled` to
// `config.intelligent_compute_enabled` too. Removed attributes are dropped by
// upgradeToCurrentSchema. See docs-internal/STATE_UPGRADER_GUIDE.md.
func GcpcloudceStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_gcp_cloud_ce", req, resp, renameCloudCESchedEnabled)
}
