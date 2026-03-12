package boolvalidators

import (
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NvmeStorageValidator() validator.Bool {
	return RequiresSiblingBoolTrue(
		"enable_fusion",
		"Fusion Must Be Enabled for NVMe Storage",
		"When 'nvme_storage_enabled' is true, 'enable_fusion' must also be set to true. "+
			"Fast NVMe instance storage requires Fusion v2 to be enabled.",
	)
}
