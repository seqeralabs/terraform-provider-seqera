package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = wspRoleValidator{}

type wspRoleValidator struct{}

func (v wspRoleValidator) Description(_ context.Context) string {
	return "validates that the workspace role is valid"
}

func (v wspRoleValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v wspRoleValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	role := req.ConfigValue.ValueString()

	validRoles := map[string]bool{
		"owner":    true,
		"admin":    true,
		"maintain": true,
		"launch":   true,
		"connect":  true,
		"view":     true,
	}

	if !validRoles[role] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Workspace Role",
			fmt.Sprintf("Workspace role must be one of: 'owner', 'admin', 'maintain', 'launch', 'connect', 'view'. Got: '%s'", role),
		)
	}
}

func WspRoleValidator() validator.String {
	return wspRoleValidator{}
}
