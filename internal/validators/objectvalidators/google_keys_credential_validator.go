package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = ObjectGoogleKeysCredentialValidatorValidator{}

type ObjectGoogleKeysCredentialValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v ObjectGoogleKeysCredentialValidatorValidator) Description(_ context.Context) string {
	return "TODO: add validator description"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v ObjectGoogleKeysCredentialValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v ObjectGoogleKeysCredentialValidatorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement objectvalidator GoogleKeysCredentialValidator logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func GoogleKeysCredentialValidator() validator.Object {
	return ObjectGoogleKeysCredentialValidatorValidator{}
}
