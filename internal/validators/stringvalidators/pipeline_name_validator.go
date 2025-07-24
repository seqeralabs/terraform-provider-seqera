package stringvalidators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringPipelineNameValidatorValidator{}

type StringPipelineNameValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringPipelineNameValidatorValidator) Description(_ context.Context) string {
	return "Pipeline name must contain a minimum of 2 and a maximum of 99 alphanumeric characters separated by dashes, dots or underscores"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringPipelineNameValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringPipelineNameValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	pipelineName := req.ConfigValue.ValueString()

	// Check length (this is also handled by UTF8LengthBetween, but we provide a clearer message)
	if len(pipelineName) < 2 || len(pipelineName) > 99 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Pipeline Name Length",
			fmt.Sprintf("Pipeline name must be between 2 and 99 characters long, got %d characters", len(pipelineName)),
		)
		return
	}

	// Check pattern: alphanumeric characters separated by dashes, dots, or underscores
	// Must start and end with alphanumeric, separators only between alphanumeric chars
	pattern := regexp.MustCompile(`^[a-zA-Z0-9]+([._-][a-zA-Z0-9]+)*$`)
	if !pattern.MatchString(pipelineName) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Pipeline Name Format",
			"Pipeline name must start and end with alphanumeric characters (a-z, A-Z, 0-9) and can contain dashes (-), dots (.), or underscores (_) as separators between alphanumeric characters. Examples: 'my-pipeline', 'api_v2.test', 'pipeline123'",
		)
		return
	}
}

func PipelineNameValidator() validator.String {
	return StringPipelineNameValidatorValidator{}
}
