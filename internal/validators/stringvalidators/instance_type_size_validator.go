package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringInstanceTypeSizeValidatorValidator{}

type StringInstanceTypeSizeValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringInstanceTypeSizeValidatorValidator) Description(_ context.Context) string {
	return "Instance type size must be one of: SMALL, MEDIUM, LARGE"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringInstanceTypeSizeValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringInstanceTypeSizeValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	validSizes := map[string]bool{
		"SMALL":  true,
		"MEDIUM": true,
		"LARGE":  true,
	}

	if !validSizes[value] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Instance Type Size",
			fmt.Sprintf("Instance type size must be one of SMALL, MEDIUM, or LARGE, got: %s", value),
		)
		return
	}
}

func InstanceTypeSizeValidator() validator.String {
	return StringInstanceTypeSizeValidatorValidator{}
}
