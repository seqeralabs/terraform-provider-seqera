package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = ObjectFusionSnapshotsUnsupportedValidator{}

// ObjectFusionSnapshotsUnsupportedValidator enforces that
// intelligent_compute_config.fusion_snapshots is not enabled on cloud platforms
// where the backend does not support it (currently Azure). The field lives on
// the shared SchedConfig schema, so it cannot be dropped for a single platform;
// this validator rejects an enabled value at plan time instead of letting the
// backend silently ignore it.
type ObjectFusionSnapshotsUnsupportedValidator struct{}

func (v ObjectFusionSnapshotsUnsupportedValidator) Description(_ context.Context) string {
	return "Validates that intelligent_compute_config.fusion_snapshots is not enabled on platforms that do not support it."
}

func (v ObjectFusionSnapshotsUnsupportedValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ObjectFusionSnapshotsUnsupportedValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var fusionSnapshots types.Bool
	snapshotsPath := req.Path.AtName("intelligent_compute_config").AtName("fusion_snapshots")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, snapshotsPath, &fusionSnapshots)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only an explicitly-enabled value is rejected; null/unknown/false is fine.
	if fusionSnapshots.IsNull() || fusionSnapshots.IsUnknown() || !fusionSnapshots.ValueBool() {
		return
	}

	resp.Diagnostics.AddAttributeError(
		snapshotsPath,
		"Unsupported fusion_snapshots",
		"Fusion snapshots are not supported on this compute environment. "+
			"Remove `fusion_snapshots` or set it to `false`.",
	)
}

// FusionSnapshotsUnsupportedValidator returns a validator that rejects an
// enabled intelligent_compute_config.fusion_snapshots.
func FusionSnapshotsUnsupportedValidator() validator.Object {
	return ObjectFusionSnapshotsUnsupportedValidator{}
}
