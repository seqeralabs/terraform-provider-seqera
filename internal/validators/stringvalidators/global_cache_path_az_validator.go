package stringvalidators

import "github.com/hashicorp/terraform-plugin-framework/schema/validator"

// GlobalCachePathAzValidator validates global_cache_path on Azure compute environments,
// which require a az:// location. See StringGlobalCachePathValidator for the rules.
func GlobalCachePathAzValidator() validator.String {
	return StringGlobalCachePathValidator{Scheme: "az"}
}
