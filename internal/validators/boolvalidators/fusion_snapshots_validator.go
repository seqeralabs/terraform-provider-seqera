package boolvalidators

import (
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func FusionSnapshotsValidator() validator.Bool {
	return RequiresSiblingBoolTrue(
		"enable_fusion",
		"Fusion Must Be Enabled for Fusion Snapshots",
		"When 'fusion_snapshots' is true, 'enable_fusion' must also be set to true. "+
			"Fusion Snapshots requires the Fusion v2 file system.",
	)
}
