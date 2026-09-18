package boolvalidators

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = RequiresSiblingStringSetValidator{}

// RequiresSiblingStringSetValidator validates that when the current bool attribute is true,
// a sibling string attribute (at the same level) is set to a non-empty value.
type RequiresSiblingStringSetValidator struct {
	SiblingAttr string
	ErrorTitle  string
	ErrorDetail string
}

func (v RequiresSiblingStringSetValidator) Description(_ context.Context) string {
	return "validates that when this attribute is true, " + v.SiblingAttr + " must be set"
}

func (v RequiresSiblingStringSetValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v RequiresSiblingStringSetValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
		return
	}

	var siblingValue types.String
	siblingPath := req.Path.ParentPath().AtName(v.SiblingAttr)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &siblingValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase (interpolated from another resource).
	if siblingValue.IsUnknown() {
		return
	}

	if siblingValue.IsNull() || strings.TrimSpace(siblingValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.ErrorTitle,
			v.ErrorDetail,
		)
	}
}

// RequiresSiblingStringSet creates a validator that checks a sibling string attribute is
// set to a non-empty value when the current attribute is true.
func RequiresSiblingStringSet(siblingAttr, errorTitle, errorDetail string) validator.Bool {
	return RequiresSiblingStringSetValidator{
		SiblingAttr: siblingAttr,
		ErrorTitle:  errorTitle,
		ErrorDetail: errorDetail,
	}
}
