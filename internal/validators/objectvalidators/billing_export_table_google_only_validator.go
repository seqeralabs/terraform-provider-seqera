package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = ObjectBillingExportTableGoogleOnlyValidator{}

// ObjectBillingExportTableGoogleOnlyValidator enforces that
// intelligent_compute_config.billing_export_table is unset on cloud platforms
// other than Google. The field names a BigQuery table holding a Cloud Billing
// export, so only the Google backend consumes it; it lives on the shared
// SchedConfig schema and so cannot be dropped for a single platform, and this
// validator rejects a configured value at plan time rather than letting the
// backend silently ignore it.
type ObjectBillingExportTableGoogleOnlyValidator struct{}

func (v ObjectBillingExportTableGoogleOnlyValidator) Description(_ context.Context) string {
	return "Validates that intelligent_compute_config.billing_export_table is unset on platforms other than Google Cloud."
}

func (v ObjectBillingExportTableGoogleOnlyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ObjectBillingExportTableGoogleOnlyValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var billingExportTable types.String
	tablePath := req.Path.AtName("intelligent_compute_config").AtName("billing_export_table")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, tablePath, &billingExportTable)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only an explicitly-configured value is rejected; null/unknown/empty is fine.
	if billingExportTable.IsNull() || billingExportTable.IsUnknown() || billingExportTable.ValueString() == "" {
		return
	}

	resp.Diagnostics.AddAttributeError(
		tablePath,
		"Unsupported billing_export_table",
		"`billing_export_table` names a BigQuery Cloud Billing export and is only supported on "+
			"Google Cloud compute environments. Remove it from this compute environment.",
	)
}

// BillingExportTableGoogleOnlyValidator returns a validator that rejects a
// configured intelligent_compute_config.billing_export_table.
func BillingExportTableGoogleOnlyValidator() validator.Object {
	return ObjectBillingExportTableGoogleOnlyValidator{}
}
