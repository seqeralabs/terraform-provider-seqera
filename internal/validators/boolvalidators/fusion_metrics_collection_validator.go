package boolvalidators

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// FusionMetricsCollectionValidator enforces that Fusion metrics collection is
// only enabled on a compute environment that has Fusion enabled. The API rejects
// the combination with a 400 ("Fusion metrics collection cannot be enabled on a
// compute environment without Fusion enabled"), so this surfaces it at plan time
// with a clear message instead of failing at apply.
//
// Only wired on the Batch compute environments, which expose `config.enable_fusion`.
// The Cloud compute environments have no Fusion toggle — Fusion is always on there —
// so there is nothing to validate against, matching how FusionSnapshotsValidator is
// already applied.
func FusionMetricsCollectionValidator() validator.Bool {
	return RequiresBoolTrueAt(
		path.Root("config").AtName("enable_fusion"),
		"Fusion Must Be Enabled for Fusion Metrics Collection",
		"When 'fusion_metrics_collection_enabled' is true, 'config.enable_fusion' must also be set to true. "+
			"Fusion metrics collection requires the Fusion v2 file system.",
	)
}
