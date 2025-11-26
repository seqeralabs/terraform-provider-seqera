package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tfTypes "github.com/seqeralabs/terraform-provider-seqera/internal/provider/types"
)

var _ validator.Object = ObjectAwsForgeValidatorValidator{}

type ObjectAwsForgeValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v ObjectAwsForgeValidatorValidator) Description(_ context.Context) string {
	return "Validates AWS Batch Forge configuration rules: 1) If forge and enable_fusion are enabled, cli_path must not be set, 2) If enable_fusion is enabled, enable_wave must be enabled"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v ObjectAwsForgeValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v ObjectAwsForgeValidatorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Skip validation if the value is null or unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Extract the AWS Batch configuration from the request
	var awsBatchConfig tfTypes.AwsBatchConfig
	diags := req.ConfigValue.As(ctx, &awsBatchConfig, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Rule 1: If forge is set (has a type) AND enable_fusion is enabled, cli_path must not be set
	forgeEnabled := awsBatchConfig.Forge != nil &&
		!awsBatchConfig.Forge.Type.IsNull() &&
		!awsBatchConfig.Forge.Type.IsUnknown() &&
		awsBatchConfig.Forge.Type.ValueString() != ""

	fusion2Enabled := !awsBatchConfig.EnableFusion.IsNull() && awsBatchConfig.EnableFusion.ValueBool()

	if forgeEnabled && fusion2Enabled {
		// Both Forge and Fusion2 are enabled - validate cli_path is not set
		if !awsBatchConfig.CliPath.IsNull() &&
			!awsBatchConfig.CliPath.IsUnknown() &&
			awsBatchConfig.CliPath.ValueString() != "" {

			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("cli_path"),
				"Invalid AWS Batch Configuration",
				"When Forge and Fusion2 (enable_fusion) are enabled, cli_path must not be set as Forge will manage the CLI path automatically.",
			)
		}
	}

	// Rule 2: If enable_fusion is set, wave must be enabled
	if !awsBatchConfig.EnableFusion.IsNull() && awsBatchConfig.EnableFusion.ValueBool() {
		// Fusion2 is enabled - validate wave is enabled
		if awsBatchConfig.EnableWave.IsNull() || !awsBatchConfig.EnableWave.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("enable_wave"),
				"Invalid AWS Batch Configuration",
				"When Fusion2 (enable_fusion) is enabled, Wave (enable_wave) must also be enabled.",
			)
		}
	}
}

func AwsForgeValidator() validator.Object {
	return ObjectAwsForgeValidatorValidator{}
}
