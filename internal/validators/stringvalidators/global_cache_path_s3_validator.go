package stringvalidators

import "github.com/hashicorp/terraform-plugin-framework/schema/validator"

// GlobalCachePathS3Validator validates global_cache_path on AWS and Seqera Compute compute environments,
// which require a s3:// location. See StringGlobalCachePathValidator for the rules.
func GlobalCachePathS3Validator() validator.String {
	return StringGlobalCachePathValidator{Scheme: "s3"}
}
