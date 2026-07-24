package listvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.List = ListStudioAllowlistRequiresPrivateValidator{}

// ListStudioAllowlistRequiresPrivateValidator enforces that allowed_user_ids is
// only set when the studio is private (is_private = true). The allow list has no
// effect on a non-private studio and the API rejects it, so this surfaces the
// error at plan time with a clear message instead of a 400 at apply.
type ListStudioAllowlistRequiresPrivateValidator struct{}

func (v ListStudioAllowlistRequiresPrivateValidator) Description(_ context.Context) string {
	return "Ensures allowed_user_ids is only set when is_private is true."
}

func (v ListStudioAllowlistRequiresPrivateValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ListStudioAllowlistRequiresPrivateValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	// No allow list configured -> nothing to enforce.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}

	var isPrivate types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("is_private"), &isPrivate)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isPrivate.IsNull() || isPrivate.IsUnknown() || !isPrivate.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"allowed_user_ids requires a private studio",
			"`allowed_user_ids` is only supported for private studios. Set `is_private = true`, or remove `allowed_user_ids`.",
		)
	}
}

// StudioAllowlistRequiresPrivateValidator ensures allowed_user_ids is only set on a private studio.
func StudioAllowlistRequiresPrivateValidator() validator.List {
	return ListStudioAllowlistRequiresPrivateValidator{}
}
