package boolvalidators

import (
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func FusionEnabledValidator() validator.Bool {
	return RequiresSiblingBoolTrue(
		"enable_wave",
		"Wave Must Be Enabled for Fusion",
		"When 'enable_fusion' is true, 'enable_wave' must also be set to true. "+
			"Fusion v2 requires Wave containers to be enabled.",
	)
}
