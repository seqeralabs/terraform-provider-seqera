package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringLineageStoreLocationValidator{}

// StringLineageStoreLocationValidator validates store_location on
// seqera_workspace_lineage_settings: not allowed in AUTO mode; in MANUAL mode
// required unless the deprecated store_name is set, and not allowed together
// with store_name. See lineage_store_common.go.
type StringLineageStoreLocationValidator struct{}

func (v StringLineageStoreLocationValidator) Description(_ context.Context) string {
	return "must be set when provisioning_mode is MANUAL (unless the deprecated store_name is set), must not be set together with store_name, and must not be set when provisioning_mode is AUTO"
}

func (v StringLineageStoreLocationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringLineageStoreLocationValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	manual, known := lineageStoreAutoCheck(ctx, req, resp)
	if !known || !manual || req.ConfigValue.IsUnknown() {
		return
	}

	storeName, diags := lineageSiblingString(ctx, req, "store_name")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || storeName.IsUnknown() {
		return
	}

	switch {
	case req.ConfigValue.IsNull() && storeName.IsNull():
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Lineage Store Location Required for MANUAL Provisioning",
			"'"+req.Path.String()+"' must be set when 'provisioning_mode' is MANUAL: "+
				"give a path within your existing bucket, for example 's3://my-bucket/lineage'.",
		)
	case !req.ConfigValue.IsNull() && !storeName.IsNull():
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting Lineage Store Attributes",
			"'"+req.Path.String()+"' and '"+lineageStorePath(req, "store_name").String()+"' cannot both be set. "+
				"'store_name' is deprecated: remove it and keep only 'store_location' "+
				"(a bucket name 'my-bucket' corresponds to 's3://my-bucket').",
		)
	}
}

func LineageStoreLocationValidator() validator.String {
	return StringLineageStoreLocationValidator{}
}
