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
	return "Validates AWS Batch Forge configuration rules: cli_path/Fusion exclusivity, DRAGEN dependencies, EFS exclusivity, EBS auto-scale dependencies."
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

	// If forge is set (has a type) AND enable_fusion is enabled, cli_path must not be set
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

	// Forge-level dependency checks. These mirror constraints the upstream
	// Seqera API enforces; emitting them at plan time gives users a faster
	// feedback loop than waiting for the API to reject the apply.
	if awsBatchConfig.Forge == nil {
		return
	}
	forge := awsBatchConfig.Forge
	forgePath := req.Path.AtName("forge")

	dragenEnabled := !forge.DragenEnabled.IsNull() && !forge.DragenEnabled.IsUnknown() && forge.DragenEnabled.ValueBool()

	// dragen_ami_id is only meaningful when dragen_enabled = true.
	if !forge.DragenAmiID.IsNull() && !forge.DragenAmiID.IsUnknown() && forge.DragenAmiID.ValueString() != "" && !dragenEnabled {
		resp.Diagnostics.AddAttributeError(
			forgePath.AtName("dragen_ami_id"),
			"Invalid AWS Batch Forge Configuration",
			"`forge.dragen_ami_id` is only applicable when `forge.dragen_enabled` is true.",
		)
	}

	// dragen_instance_type is only meaningful when dragen_enabled = true.
	if !forge.DragenInstanceType.IsNull() && !forge.DragenInstanceType.IsUnknown() && forge.DragenInstanceType.ValueString() != "" && !dragenEnabled {
		resp.Diagnostics.AddAttributeError(
			forgePath.AtName("dragen_instance_type"),
			"Invalid AWS Batch Forge Configuration",
			"`forge.dragen_instance_type` is only applicable when `forge.dragen_enabled` is true.",
		)
	}

	// efs_create and efs_id are mutually exclusive — set one or the other.
	efsCreate := !forge.EfsCreate.IsNull() && !forge.EfsCreate.IsUnknown() && forge.EfsCreate.ValueBool()
	efsIDSet := !forge.EfsID.IsNull() && !forge.EfsID.IsUnknown() && forge.EfsID.ValueString() != ""
	if efsCreate && efsIDSet {
		resp.Diagnostics.AddAttributeError(
			forgePath.AtName("efs_create"),
			"Invalid AWS Batch Forge Configuration",
			"`forge.efs_create` and `forge.efs_id` are mutually exclusive — set one or the other.",
		)
	}

	// ebs_block_size requires ebs_auto_scale = true (both are deprecated, but
	// we still validate while they exist).
	ebsAutoScale := !forge.EbsAutoScale.IsNull() && !forge.EbsAutoScale.IsUnknown() && forge.EbsAutoScale.ValueBool()
	ebsBlockSizeSet := !forge.EbsBlockSize.IsNull() && !forge.EbsBlockSize.IsUnknown() && forge.EbsBlockSize.ValueInt32() != 0
	if ebsBlockSizeSet && !ebsAutoScale {
		resp.Diagnostics.AddAttributeError(
			forgePath.AtName("ebs_block_size"),
			"Invalid AWS Batch Forge Configuration",
			"`forge.ebs_block_size` is only applicable when `forge.ebs_auto_scale` is true.",
		)
	}
}

func AwsForgeValidator() validator.Object {
	return ObjectAwsForgeValidatorValidator{}
}
