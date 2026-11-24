package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Shared logic for the lineage store validators (store_location and the
// deprecated store_name on seqera_workspace_lineage_settings).
//
// The lineage store is chosen by Platform in AUTO mode, so neither field may be
// set there. In MANUAL mode the store must be given: store_location is the
// supported field, store_name is still accepted on its own so existing
// configurations keep working, and setting both is ambiguous.

// lineageSiblingString reads a sibling string attribute of the validated one.
func lineageSiblingString(ctx context.Context, req validator.StringRequest, name string) (types.String, diag.Diagnostics) {
	var v types.String
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName(name), &v)
	return v, diags
}

// lineageStoreAutoCheck reports an error when the field is set outside MANUAL
// mode. It returns manual and known for the caller's MANUAL-mode checks; known
// is false when the mode is unknown and no further check should run.
func lineageStoreAutoCheck(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) (manual bool, known bool) {
	mode, diags := lineageSiblingString(ctx, req, "provisioning_mode")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || mode.IsUnknown() {
		return false, false
	}

	manual = !mode.IsNull() && mode.ValueString() == "MANUAL"
	if !manual && !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Lineage Field Not Allowed for AUTO Provisioning",
			"'"+req.Path.String()+"' can only be set when 'provisioning_mode' is MANUAL; "+
				"in AUTO mode Platform chooses the lineage store.",
		)
	}
	return manual, true
}

func lineageStorePath(req validator.StringRequest, name string) path.Path {
	return req.Path.ParentPath().AtName(name)
}
