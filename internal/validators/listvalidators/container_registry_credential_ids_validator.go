package listvalidators

import (
	"regexp"

	framework_listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	custom_stringvalidators "github.com/seqeralabs/terraform-provider-seqera/internal/validators/stringvalidators"
)

var containerRegistryCredentialIDPattern = regexp.MustCompile(`^[0-9A-Za-z]{1,22}$`)

// ContainerRegistryCredentialIdsValidator validates container_reg_ids values.
func ContainerRegistryCredentialIdsValidator() validator.List {
	return framework_listvalidator.ValueStringsAre(
		custom_stringvalidators.NotNull(),
		stringvalidator.RegexMatches(
			containerRegistryCredentialIDPattern,
			"must be a Seqera Platform credential ID of 1 to 22 alphanumeric characters",
		),
	)
}
