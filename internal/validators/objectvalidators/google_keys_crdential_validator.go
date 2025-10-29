package objectvalidators

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ validator.Object = ObjectGoogleKeysCrdentialValidatorValidator{}

type ObjectGoogleKeysCrdentialValidatorValidator struct{}

// Description describes the validation in plain text formatting.
func (v ObjectGoogleKeysCrdentialValidatorValidator) Description(_ context.Context) string {
	return "validates that the data field contains a valid Google service account key JSON"
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

	// Check if data field exists and is not null
	dataAttr, exists := attrs["data"]
	if !exists || dataAttr.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("data"),
			"Missing GCP Data Parameter",
			"GCP credentials requires 'data' field to be set in 'keys' and not empty",
		)
		return
	}

	// Get the string value
	dataValue := dataAttr
	stringValue, ok := dataValue.(basetypes.StringValue)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("data"),
			"Invalid Data Type",
			"The 'data' field must be a string containing valid JSON",
		)
		return
	}

	if stringValue.IsUnknown() || stringValue.IsNull() {
		return
	}

	jsonData := stringValue.ValueString()
	if jsonData == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("data"),
			"Empty GCP Data",
			"The 'data' field cannot be empty",
		)
		return
	}

	// Validate it's valid JSON
	var serviceAccount map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &serviceAccount); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("data"),
			"Invalid JSON",
			fmt.Sprintf("The 'data' field must contain valid JSON. Error: %s\n\nTip: Use file() function to read from a file:\n  keys = {\n    data = file(\"${path.module}/service-account-key.json\")\n  }", err.Error()),
		)
		return
	}

	// Validate required Google service account key fields
	requiredFields := []string{"type", "project_id", "private_key", "client_email"}
	var missingFields []string

	for _, field := range requiredFields {
		if _, exists := serviceAccount[field]; !exists {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("data"),
			"Invalid Google Service Account Key",
			fmt.Sprintf("The service account key JSON is missing required fields: %v\n\nRequired fields: %v\n\nTip: Download the key from Google Cloud Console > IAM & Admin > Service Accounts", missingFields, requiredFields),
		)
		return
	}

	// Validate type field is "service_account"
	if typeVal, ok := serviceAccount["type"].(string); ok {
		if typeVal != "service_account" {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("data"),
				"Invalid Key Type",
				fmt.Sprintf("The 'type' field must be 'service_account', got: '%s'", typeVal),
			)
			return
		}
	}
}

func GoogleKeysCrdentialValidator() validator.Object {
	return ObjectGoogleKeysCrdentialValidatorValidator{}
}
