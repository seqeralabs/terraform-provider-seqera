package stringvalidators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringLabelNameValidatorValidator{}

type StringLabelNameValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringLabelNameValidatorValidator) Description(_ context.Context) string {
	return "Label name must contain a minimum of 1 and a maximum of 39 alphanumeric characters separated by dashes or underscores"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringLabelNameValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringLabelNameValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Check length (this is also handled by UTF8LengthBetween, but we provide a clearer message)
	if len(value) < 1 || len(value) > 39 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Label Name Length",
			fmt.Sprintf("Label name must be between 1 and 39 characters long, got %d characters", len(value)),
		)
		return
	}

	// Check pattern: alphanumeric characters separated by dashes or underscores
	// Must start and end with alphanumeric, separators only between alphanumeric chars
	pattern := regexp.MustCompile(`^[a-zA-Z0-9]+([_-][a-zA-Z0-9]+)*$`)
	if !pattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Label Name Format",
			"Label name must start and end with alphanumeric characters (a-z, A-Z, 0-9) and can contain dashes (-) or underscores (_) as separators between alphanumeric characters. Examples: 'my-label', 'test_123', 'environment'",
		)
		return
	}
}

func LabelNameValidator() validator.String {
	return StringLabelNameValidatorValidator{}
}
