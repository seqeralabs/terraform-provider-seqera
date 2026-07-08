package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// AwsbatchceStateUpgraderV0 migrates seqera_aws_batch_ce state from schema
// version 0 to the current schema: it renames the misspelled
// `nvnme_storage_enabled` flag, then re-decodes against the current schema,
// dropping any attribute the schema no longer defines. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func AwsbatchceStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_aws_batch_ce", req, resp, renameNvmeStorageFlag)
}

// AwsbatchceStateUpgraderV1 migrates seqera_aws_batch_ce state by renaming the
// root `compute_env_id` attribute to `id`, then re-decoding against the current
// schema.
func AwsbatchceStateUpgraderV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_aws_batch_ce", req, resp, renameComputeEnvIDToID)
}
