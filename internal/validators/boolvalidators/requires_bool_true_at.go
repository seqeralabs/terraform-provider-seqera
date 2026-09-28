package boolvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = RequiresBoolTrueAtValidator{}

// RequiresBoolTrueAtValidator validates that when the current bool attribute is
// true, a bool attribute at an absolute path elsewhere in the config is also true.
//
// This is the cross-level counterpart to RequiresSiblingBoolTrueValidator, which
// resolves its target relative to the current attribute and so can only reach
// siblings. Use this when the two attributes sit at different depths — e.g. a
// root-level flag that depends on something nested inside `config`.
type RequiresBoolTrueAtValidator struct {
	TargetPath  path.Path
	ErrorTitle  string
	ErrorDetail string
}

func (v RequiresBoolTrueAtValidator) Description(_ context.Context) string {
	return "validates that when this attribute is true, " + v.TargetPath.String() + " must also be true"
}

func (v RequiresBoolTrueAtValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v RequiresBoolTrueAtValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
		return
	}

	var targetValue types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, v.TargetPath, &targetValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Not yet known at plan time (e.g. interpolated from another resource) —
	// erroring would be a false positive, so leave it to the API.
	if targetValue.IsUnknown() {
		return
	}

	if targetValue.IsNull() || !targetValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.ErrorTitle,
			v.ErrorDetail,
		)
	}
}

// RequiresBoolTrueAt creates a validator that checks the bool at targetPath is
// true when the current attribute is true.
func RequiresBoolTrueAt(targetPath path.Path, errorTitle, errorDetail string) validator.Bool {
	return RequiresBoolTrueAtValidator{
		TargetPath:  targetPath,
		ErrorTitle:  errorTitle,
		ErrorDetail: errorDetail,
	}
}
