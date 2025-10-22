package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = orgRoleValidator{}

type orgRoleValidator struct{}

func (v orgRoleValidator) Description(_ context.Context) string {
	return "validates that the organization role is valid"
}

func (v orgRoleValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v orgRoleValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	role := req.ConfigValue.ValueString()

	validRoles := map[string]bool{
		"owner":        true,
		"member":       true,
		"collaborator": true,
	}

	if !validRoles[role] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Organization Role",
			fmt.Sprintf("Organization role must be one of: 'owner', 'member', 'collaborator'. Got: '%s'", role),
		)
	}
}

func OrgRoleValidator() validator.String {
	return orgRoleValidator{}
}
