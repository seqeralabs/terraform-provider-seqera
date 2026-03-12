package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tfTypes "github.com/seqeralabs/terraform-provider-seqera/internal/provider/types"
)

var _ validator.Object = ObjectFargateHeadValidator{}

type ObjectFargateHeadValidator struct{}

func (v ObjectFargateHeadValidator) Description(_ context.Context) string {
	return "validates Fargate head job requirements: Fusion v2 and Spot provisioning required, EFS and FSx not compatible"
}

func (v ObjectFargateHeadValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ObjectFargateHeadValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var forge tfTypes.ForgeConfig
	diags := req.ConfigValue.As(ctx, &forge, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip if fargate_head_enabled is not set or false
	if forge.FargateHeadEnabled.IsNull() || forge.FargateHeadEnabled.IsUnknown() || !forge.FargateHeadEnabled.ValueBool() {
		return
	}

	// Fargate requires Spot provisioning model
	if !forge.Type.IsNull() && !forge.Type.IsUnknown() && forge.Type.ValueString() != "SPOT" {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("fargate_head_enabled"),
			"Fargate Requires Spot Provisioning",
			"When 'fargate_head_enabled' is true, forge 'type' must be 'SPOT'. "+
				"Fargate for head jobs requires the Spot provisioning model.",
		)
	}

	// Fargate requires Fusion v2 - reach up to parent AwsBatchConfig
	var fusionValue types.Bool
	fusionPath := req.Path.ParentPath().AtName("enable_fusion")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, fusionPath, &fusionValue)...)
	if !resp.Diagnostics.HasError() && !fusionValue.IsUnknown() {
		if fusionValue.IsNull() || !fusionValue.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("fargate_head_enabled"),
				"Fargate Requires Fusion v2",
				"When 'fargate_head_enabled' is true, 'enable_fusion' must also be set to true. "+
					"Fargate for head jobs requires the Fusion v2 file system.",
			)
		}
	}

	// EFS create not compatible
	if !forge.EfsCreate.IsNull() && !forge.EfsCreate.IsUnknown() && forge.EfsCreate.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("efs_create"),
			"EFS Not Compatible with Fargate Head Job",
			"When 'fargate_head_enabled' is true, 'efs_create' cannot be enabled. "+
				"EFS file systems are not compatible with Fargate head jobs.",
		)
	}

	// EFS ID not compatible
	if !forge.EfsID.IsNull() && !forge.EfsID.IsUnknown() && forge.EfsID.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("efs_id"),
			"EFS Not Compatible with Fargate Head Job",
			"When 'fargate_head_enabled' is true, 'efs_id' cannot be set. "+
				"EFS file systems are not compatible with Fargate head jobs.",
		)
	}

	// FSx not compatible
	if !forge.FsxName.IsNull() && !forge.FsxName.IsUnknown() && forge.FsxName.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("fsx_name"),
			"FSx Not Compatible with Fargate Head Job",
			"When 'fargate_head_enabled' is true, 'fsx_name' cannot be set. "+
				"FSx for Lustre file systems are not compatible with Fargate head jobs.",
		)
	}
}

func FargateHeadValidator() validator.Object {
	return ObjectFargateHeadValidator{}
}
