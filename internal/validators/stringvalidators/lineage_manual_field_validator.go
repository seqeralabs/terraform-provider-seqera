package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringLineageManualFieldValidator{}

// StringLineageManualFieldValidator validates that a lineage field (store_name,
// topic_arn) is supplied only when provisioning_mode is MANUAL, and is supplied
// whenever provisioning_mode is MANUAL. A null provisioning_mode means AUTO.
type StringLineageManualFieldValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringLineageManualFieldValidator) Description(_ context.Context) string {
	return "must be set when provisioning_mode is MANUAL and must not be set when provisioning_mode is AUTO"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringLineageManualFieldValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v StringLineageManualFieldValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Unknown values may resolve to anything at apply time.
	if req.ConfigValue.IsUnknown() {
		return
	}

	var mode types.String
	siblingPath := req.Path.ParentPath().AtName("provisioning_mode")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &mode)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if mode.IsUnknown() {
		return
	}

	manual := !mode.IsNull() && mode.ValueString() == "MANUAL"

	switch {
	case manual && req.ConfigValue.IsNull():
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Lineage Field Required for MANUAL Provisioning",
			"'"+req.Path.String()+"' must be set when 'provisioning_mode' is MANUAL.",
		)
	case !manual && !req.ConfigValue.IsNull():
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Lineage Field Not Allowed for AUTO Provisioning",
			"'"+req.Path.String()+"' can only be set when 'provisioning_mode' is MANUAL; "+
				"in AUTO mode the backend ignores this value.",
		)
	}
}

// LineageManualFieldValidator returns a validator ensuring a lineage field is
// set if and only if provisioning_mode is MANUAL.
func LineageManualFieldValidator() validator.String {
	return StringLineageManualFieldValidator{}
}
