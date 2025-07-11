package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = ObjectGoogleKeysCrdentialValidatorValidator{}

type ObjectGoogleKeysCrdentialValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v ObjectGoogleKeysCrdentialValidatorValidator) Description(_ context.Context) string {
	return "validates that the data field is not null for Google credentials"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v ObjectGoogleKeysCrdentialValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v ObjectGoogleKeysCrdentialValidatorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Skip validation if object is null or unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	// Return error if data field is empty or null in GoogleSecurityKeys.

	if dataAttr, exists := attrs["data"]; exists {
		if dataAttr.IsNull() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("data"),
				"Missing GCP Data Parameter",
				"GCP credentials requires 'data' field to be set in 'keys' and not empty",
			)
		}
	}

}

func GoogleKeysCrdentialValidator() validator.Object {
	return ObjectGoogleKeysCrdentialValidatorValidator{}
}
