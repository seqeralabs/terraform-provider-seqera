package boolvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = RequiresSiblingBoolTrueValidator{}

// RequiresSiblingBoolTrueValidator validates that when the current bool attribute is true,
// a sibling bool attribute (at the same level) must also be true.
type RequiresSiblingBoolTrueValidator struct {
	SiblingAttr string
	ErrorTitle  string
	ErrorDetail string
}

func (v RequiresSiblingBoolTrueValidator) Description(_ context.Context) string {
	return "validates that when this attribute is true, " + v.SiblingAttr + " must also be true"
}

func (v RequiresSiblingBoolTrueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v RequiresSiblingBoolTrueValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
		return
	}

	var siblingValue types.Bool
	siblingPath := req.Path.ParentPath().AtName(v.SiblingAttr)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &siblingValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if siblingValue.IsUnknown() {
		return
	}

	if siblingValue.IsNull() || !siblingValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			v.ErrorTitle,
			v.ErrorDetail,
		)
	}
}

// RequiresSiblingBoolTrue creates a validator that checks a sibling bool attribute is true
// when the current attribute is true.
func RequiresSiblingBoolTrue(siblingAttr, errorTitle, errorDetail string) validator.Bool {
	return RequiresSiblingBoolTrueValidator{
		SiblingAttr: siblingAttr,
		ErrorTitle:  errorTitle,
		ErrorDetail: errorDetail,
	}
}
