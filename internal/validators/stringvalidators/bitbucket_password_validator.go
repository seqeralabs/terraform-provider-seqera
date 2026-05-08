package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringBitbucketPasswordValidatorValidator{}

type StringBitbucketPasswordValidatorValidator struct{}

func (v StringBitbucketPasswordValidatorValidator) Description(_ context.Context) string {
	return "validates that `password` and `token` are not both set on a Bitbucket credential"
}

func (v StringBitbucketPasswordValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringBitbucketPasswordValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateBitbucketKeysSibling(ctx, req, resp, "token")
}

func BitbucketPasswordValidator() validator.String {
	return StringBitbucketPasswordValidatorValidator{}
}

// validateBitbucketKeysSibling errors when the current attribute and its sibling
// are both set. `password` and `token` are mutually exclusive on the API.
func validateBitbucketKeysSibling(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse, sibling string) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}

	var siblingValue types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName(sibling), &siblingValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if siblingValue.IsUnknown() || siblingValue.IsNull() || siblingValue.ValueString() == "" {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Conflicting Bitbucket Credentials",
		"Only one of `password` or `token` may be set.",
	)
}
