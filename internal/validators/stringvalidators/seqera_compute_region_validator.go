package stringvalidators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringSeqeraComputeRegionValidatorValidator{}

type StringSeqeraComputeRegionValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringSeqeraComputeRegionValidatorValidator) Description(_ context.Context) string {
	return "Region must be a valid AWS region (e.g., us-east-1, eu-west-1, ap-southeast-2)"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringSeqeraComputeRegionValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringSeqeraComputeRegionValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// AWS region format: <prefix>-<direction>-<number>
	// Examples: us-east-1, eu-west-2, ap-southeast-1, sa-east-1, ca-central-1, me-south-1, af-south-1
	pattern := regexp.MustCompile(`^(us|eu|ap|sa|ca|me|af|il)-(east|west|north|south|central|northeast|southeast|northwest|southwest)-\d+$`)
	if !pattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Region",
			fmt.Sprintf("Region must be a valid region format (e.g., us-east-1, eu-west-1, ap-southeast-2), got: %s", value),
		)
		return
	}
}

func SeqeraComputeRegionValidator() validator.String {
	return StringSeqeraComputeRegionValidatorValidator{}
}
