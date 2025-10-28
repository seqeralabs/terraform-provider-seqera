package boolvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = BoolLabelIsDefaultValidatorValidator{}

type BoolLabelIsDefaultValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v BoolLabelIsDefaultValidatorValidator) Description(_ context.Context) string {
	return "validates that label is_default can only be true when resource is true"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v BoolLabelIsDefaultValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v BoolLabelIsDefaultValidatorValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// Skip validation if value is null, unknown, or false
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
		return
	}

	// At this point, is_default is true - validate that resource is also true
	var resourceValue types.Bool
	resourcePath := req.Path.ParentPath().AtName("resource")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, resourcePath, &resourceValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown resource values during plan phase
	if resourceValue.IsUnknown() {
		return
	}

	// Require resource to be explicitly set to true
	if resourceValue.IsNull() || !resourceValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Label Configuration",
			"The 'is_default' attribute can only be set to true when 'resource' is true. Resource labels (resource=true) can be marked as default, but non-resource labels cannot.",
		)
	}
}

func LabelIsDefaultValidator() validator.Bool {
	return BoolLabelIsDefaultValidatorValidator{}
}
