package stateupgraders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ContainerregistrycredentialStateUpgraderV0 upgrades seqera_container_registry_credential state to the current schema by re-decoding prior state
// against it, dropping any attribute the schema no longer defines. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func ContainerregistrycredentialStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_container_registry_credential", req, resp, nil)
}
