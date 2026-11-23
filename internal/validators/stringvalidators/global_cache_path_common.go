package stringvalidators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringGlobalCachePathValidator{}

// StringGlobalCachePathValidator mirrors the offline checks the Platform applies to a
// compute environment's Nextflow Global Cache location (GlobalCacheLocationValidator):
//
//   - the location may only be set when global_cache_enabled is true;
//   - it must be an object storage URI using the scheme of the compute environment's
//     platform (s3:// for AWS and Seqera Compute, gs:// for Google, az:// for Azure)
//     and must name a bucket;
//   - it must already be in the canonical form the Platform stores. The Platform
//     collapses repeated slashes and strips trailing slashes, so a non-canonical value
//     would come back different from the configuration and cause an "inconsistent
//     result after apply" error, followed by a forced replacement on the next plan.
type StringGlobalCachePathValidator struct {
	// Scheme is the object storage scheme required by the compute environment platform, without "://".
	Scheme string
}

func (v StringGlobalCachePathValidator) Description(_ context.Context) string {
	return fmt.Sprintf("global_cache_path can only be set when global_cache_enabled is true, and must be a canonical %s:// URI naming a bucket", v.Scheme)
}

func (v StringGlobalCachePathValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringGlobalCachePathValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	location := req.ConfigValue.ValueString()

	var enabled types.Bool
	siblingPath := req.Path.ParentPath().AtName("global_cache_enabled")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &enabled)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !enabled.IsUnknown() && (enabled.IsNull() || !enabled.ValueBool()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Global Cache Not Enabled",
			"'global_cache_path' can only be set when 'global_cache_enabled' is true. "+
				"Set 'global_cache_enabled = true', or remove 'global_cache_path'.",
		)
	}

	idx := strings.Index(location, "://")
	if idx <= 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Global Cache Location",
			fmt.Sprintf("'global_cache_path' must be an object storage URI such as '%s://my-bucket/cache', got '%s'.", v.Scheme, location),
		)
		return
	}

	if scheme := location[:idx]; scheme != v.Scheme {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Global Cache Location Scheme",
			fmt.Sprintf("'global_cache_path' uses the '%s://' scheme, which does not match the '%s://' storage of this compute environment.", scheme, v.Scheme),
		)
		return
	}

	normalised := normaliseGlobalCachePath(location)
	if normalised == v.Scheme+"://" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Global Cache Location",
			fmt.Sprintf("'global_cache_path' must name a bucket, got '%s'.", location),
		)
		return
	}
	if normalised != location {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Non-canonical Global Cache Location",
			fmt.Sprintf("The Platform stores the global cache location without repeated or trailing slashes, "+
				"so '%s' would be saved as '%s' and cause a perpetual diff. Use '%s' instead.", location, normalised, normalised),
		)
	}
}

// normaliseGlobalCachePath reproduces GlobalCacheLocationValidator.normalise on the Platform:
// repeated slashes after the scheme are collapsed and trailing slashes are removed.
func normaliseGlobalCachePath(location string) string {
	idx := strings.Index(location, "://")
	if idx < 0 {
		return location
	}
	scheme, rest := location[:idx], location[idx+3:]
	for strings.Contains(rest, "//") {
		rest = strings.ReplaceAll(rest, "//", "/")
	}
	rest = strings.TrimRight(rest, "/")
	return scheme + "://" + rest
}
