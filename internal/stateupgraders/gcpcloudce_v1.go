package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// GcpcloudceStateUpgraderV1 migrates seqera_gcp_cloud_ce state from schema version 1 to
// the current schema: `config.sched_enabled` was renamed to
// `config.intelligent_compute_enabled` to match seqera_aws_cloud_ce. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func GcpcloudceStateUpgraderV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_gcp_cloud_ce", req, resp, renameCloudCESchedEnabled)
}
