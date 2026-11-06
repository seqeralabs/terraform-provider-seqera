package stringvalidators

import "github.com/hashicorp/terraform-plugin-framework/schema/validator"

// GlobalCachePathGsValidator validates global_cache_path on Google compute environments,
// which require a gs:// location. See StringGlobalCachePathValidator for the rules.
func GlobalCachePathGsValidator() validator.String {
	return StringGlobalCachePathValidator{Scheme: "gs"}
}
