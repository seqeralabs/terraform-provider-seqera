package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = managedIdentityPlatformValidator{}

type managedIdentityPlatformValidator struct{}

func (v managedIdentityPlatformValidator) Description(_ context.Context) string {
	return "validates that the platform value is a valid grid computing platform"
}

func (v managedIdentityPlatformValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v managedIdentityPlatformValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	platform := req.ConfigValue.ValueString()

	// Valid grid computing platforms
	validPlatforms := map[string]bool{
		"altair-platform": true,
		"lsf-platform":    true,
		"moab-platform":   true,
		"slurm-platform":  true,
		"uge-platform":    true,
	}

	if !validPlatforms[platform] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Grid Platform",
			fmt.Sprintf("Platform '%s' is not a valid grid computing platform. Valid platforms are: altair-platform, lsf-platform, moab-platform, slurm-platform, uge-platform", platform),
		)
	}
}

func ManagedIdentityPlatformValidator() validator.String {
	return managedIdentityPlatformValidator{}
}
