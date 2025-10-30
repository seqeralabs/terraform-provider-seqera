package boolvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = BoolWaveEnabledValidator{}

type BoolWaveEnabledValidator struct{}

// Description describes the validation in plain text formatting.
func (v BoolWaveEnabledValidator) Description(_ context.Context) string {
	return "validates that when enable_wave is set, enable_fusion must be explicitly configured"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v BoolWaveEnabledValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v BoolWaveEnabledValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// Skip validation if enable_wave is null, unknown, or false
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
		return
	}

	// At this point, enable_wave is true - validate that enable_fusion is explicitly set
	var fusionValue types.Bool
	fusionPath := req.Path.ParentPath().AtName("enable_fusion")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, fusionPath, &fusionValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown enable_fusion values during plan phase
	if fusionValue.IsUnknown() {
		return
	}

	// Require enable_fusion to be explicitly set (not null) when wave is enabled
	if fusionValue.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Configuration",
			"When 'enable_wave' is true, 'enable_fusion' must be explicitly set to either true or false. "+
				"Wave containers work with or without Fusion2, but you must explicitly configure this setting. "+
				"Note: If you enable Fusion2 (enable_fusion=true), Wave must also be enabled (enable_wave=true).",
		)
	}
}

func WaveEnabledValidator() validator.Bool {
	return BoolWaveEnabledValidator{}
}
