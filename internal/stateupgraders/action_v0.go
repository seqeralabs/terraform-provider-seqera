package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ActionStateUpgraderV0 migrates seqera_action state from schema version 0 to the
// current schema. It renames the misspelled `nvnme_storage_enabled` flag in the
// embedded compute-env config, then re-decodes against the current schema,
// dropping any attribute the schema no longer defines. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func ActionStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_action", req, resp, renameNvmeStorageFlag)
}
