package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// GitlabcredentialStateUpgraderV0 upgrades seqera_gitlab_credential state to the current schema by re-decoding prior state
// against it, dropping any attribute the schema no longer defines. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func GitlabcredentialStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_gitlab_credential", req, resp, nil)
}
