package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringLabelValueResourceValidatorValidator{}

type StringLabelValueResourceValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringLabelValueResourceValidatorValidator) Description(_ context.Context) string {
	return "validates that value must be set when resource is true"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringLabelValueResourceValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringLabelValueResourceValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Get the resource field value
	var resourceValue types.Bool
	resourcePath := req.Path.ParentPath().AtName("resource")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, resourcePath, &resourceValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase
	if resourceValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	resourceIsTrue := !resourceValue.IsNull() && resourceValue.ValueBool()
	valueIsEmpty := req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == ""

	// Rule 1: If resource=true, value must be set
	if resourceIsTrue && valueIsEmpty {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Attribute",
			"The 'value' attribute is required when 'resource' is true. Resource labels must have a value assigned to them.",
		)
		return
	}

	// Rule 2: If value is set, resource must be true
	if !valueIsEmpty && !resourceIsTrue {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Label Configuration",
			"The 'value' attribute can only be set when 'resource' is true. Only resource labels (resource=true) can have values assigned to them.",
		)
		return
	}

	// Reuse LabelNameValidator for format validation
	if !valueIsEmpty {
		nameValidator := LabelNameValidator()
		nameValidator.ValidateString(ctx, req, resp)
	}
}

func LabelValueResourceValidator() validator.String {
	return StringLabelValueResourceValidatorValidator{}
}
