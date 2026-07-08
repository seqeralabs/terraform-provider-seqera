package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// AwscomputeenvStateUpgraderV0 migrates seqera_aws_compute_env state from schema
// version 0 to the current schema: it renames the misspelled
// `nvnme_storage_enabled` flag, then re-decodes against the current schema,
// dropping any attribute the schema no longer defines. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func AwscomputeenvStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_aws_compute_env", req, resp, renameNvmeStorageFlag)
}

// AwscomputeenvStateUpgraderV1 migrates seqera_aws_compute_env state by renaming
// the root `compute_env_id` attribute to `id`, then re-decoding against the
// current schema.
func AwscomputeenvStateUpgraderV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_aws_compute_env", req, resp, renameComputeEnvIDToID)
}
