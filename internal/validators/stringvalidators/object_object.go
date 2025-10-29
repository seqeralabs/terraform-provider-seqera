package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringObjectObjectValidator{}

type StringObjectObjectValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringObjectObjectValidator) Description(_ context.Context) string {
	return "TODO: add validator description"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringObjectObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringObjectObjectValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement stringvalidator ObjectObject logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func ObjectObject() validator.String {
	return StringObjectObjectValidator{}
}
