package stringvalidators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringComputeEnvNameValidator{}

type StringComputeEnvNameValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringComputeEnvNameValidator) Description(_ context.Context) string {
	return "Compute environment name must contain 1 to 100 alphanumeric characters, dashes, or underscores"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringComputeEnvNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringComputeEnvNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	computeEnvName := req.ConfigValue.ValueString()

	// Check length
	if len(computeEnvName) < 1 || len(computeEnvName) > 100 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Compute Environment Name Length",
			fmt.Sprintf("Compute environment name must be between 1 and 100 characters long, got %d characters", len(computeEnvName)),
		)
		return
	}

	// Check pattern: alphanumeric characters, dashes, or underscores
	// This matches the Seqera Platform requirements for compute environment names
	pattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !pattern.MatchString(computeEnvName) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Compute Environment Name Format",
			"Compute environment name must contain only alphanumeric characters (a-z, A-Z, 0-9), dashes (-), or underscores (_). Examples: 'my-compute-env', 'prod_gcp_batch', 'aws-batch-01'",
		)
		return
	}
}

func ComputeEnvNameValidator() validator.String {
	return StringComputeEnvNameValidator{}
}
