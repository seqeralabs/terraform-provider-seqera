package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringFusionVersionValidator{}

// StringFusionVersionValidator validates that fusion_version is only supplied
// when override_nextflow_fusion_version is true. The backend ignores
// fusionVersion otherwise, so a set value would silently have no effect.
type StringFusionVersionValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringFusionVersionValidator) Description(_ context.Context) string {
	return "fusion_version can only be set when override_nextflow_fusion_version is true"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringFusionVersionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v StringFusionVersionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Nothing to check if no Fusion version is configured.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var override types.Bool
	siblingPath := req.Path.ParentPath().AtName("override_nextflow_fusion_version")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &override)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase (interpolated from another resource).
	if override.IsUnknown() {
		return
	}

	if override.IsNull() || !override.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Fusion Version Override Not Enabled",
			"'fusion_version' can only be set when 'override_nextflow_fusion_version' is true; "+
				"otherwise the Fusion version bundled with the Nextflow launcher container is used and this value is ignored.",
		)
	}
}

// FusionVersionValidator returns a validator ensuring fusion_version is only set
// when override_nextflow_fusion_version is true.
func FusionVersionValidator() validator.String {
	return StringFusionVersionValidator{}
}
