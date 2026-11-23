package boolvalidators

import "github.com/hashicorp/terraform-plugin-framework/schema/validator"

// GlobalCacheEnabledValidator ensures global_cache_path is set when global_cache_enabled is true.
// The Platform treats the toggle and the location as a single unit and rejects an enabled
// global cache without a location. The reverse direction (a location without the toggle) is
// enforced by the global_cache_path validators.
func GlobalCacheEnabledValidator() validator.Bool {
	return RequiresSiblingStringSet(
		"global_cache_path",
		"Missing Global Cache Location",
		"'global_cache_path' is required when 'global_cache_enabled' is true. "+
			"Set it to an object storage folder, for example 's3://my-bucket/nextflow-cache'.",
	)
}
