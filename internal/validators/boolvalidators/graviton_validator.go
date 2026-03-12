package boolvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = BoolGravitonValidator{}

type BoolGravitonValidator struct{}

func (v BoolGravitonValidator) Description(_ context.Context) string {
	return "validates that when arm64_enabled is true, Fargate, Wave, and Fusion v2 are all enabled"
}

func (v BoolGravitonValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v BoolGravitonValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// Skip if arm64_enabled is null, unknown, or false
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
		return
	}

	// arm64_enabled is inside ForgeConfig, so ParentPath = forge object
	forgePath := req.Path.ParentPath()

	// Check fargate_head_enabled (sibling in ForgeConfig)
	var fargateValue types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, forgePath.AtName("fargate_head_enabled"), &fargateValue)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !fargateValue.IsUnknown() && (fargateValue.IsNull() || !fargateValue.ValueBool()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Graviton Requires Fargate",
			"When 'arm64_enabled' is true, 'fargate_head_enabled' must also be set to true. "+
				"Graviton (ARM64) CPU architecture requires Fargate for head jobs.",
		)
	}

	// Check enable_fusion (on parent AwsBatchConfig)
	configPath := forgePath.ParentPath()

	var fusionValue types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, configPath.AtName("enable_fusion"), &fusionValue)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !fusionValue.IsUnknown() && (fusionValue.IsNull() || !fusionValue.ValueBool()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Graviton Requires Fusion v2",
			"When 'arm64_enabled' is true, 'enable_fusion' must also be set to true. "+
				"Graviton (ARM64) CPU architecture requires the Fusion v2 file system.",
		)
	}

	// Check enable_wave (on parent AwsBatchConfig)
	var waveValue types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, configPath.AtName("enable_wave"), &waveValue)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !waveValue.IsUnknown() && (waveValue.IsNull() || !waveValue.ValueBool()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Graviton Requires Wave",
			"When 'arm64_enabled' is true, 'enable_wave' must also be set to true. "+
				"Graviton (ARM64) CPU architecture requires the Wave containers service.",
		)
	}
}

func GravitonValidator() validator.Bool {
	return BoolGravitonValidator{}
}
