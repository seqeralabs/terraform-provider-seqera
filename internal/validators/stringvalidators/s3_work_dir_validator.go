package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringS3WorkDirValidatorValidator{}

type StringS3WorkDirValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringS3WorkDirValidatorValidator) Description(_ context.Context) string {
	return "TODO: add validator description"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringS3WorkDirValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringS3WorkDirValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement stringvalidator S3WorkDirValidator logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func S3WorkDirValidator() validator.String {
	return StringS3WorkDirValidatorValidator{}
}
