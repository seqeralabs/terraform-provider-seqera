package listvalidators

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var containerRegistryCredentialIDPattern = regexp.MustCompile(`^[0-9A-Za-z]{1,22}$`)

var _ validator.List = ListContainerRegistryCredentialIdsValidator{}

// ListContainerRegistryCredentialIdsValidator ensures the list contains Seqera
// Platform credential IDs encoded using CompactUuid's Base62 format.
type ListContainerRegistryCredentialIdsValidator struct{}

func (v ListContainerRegistryCredentialIdsValidator) Description(_ context.Context) string {
	return "values must be Seqera Platform container registry credential IDs"
}

func (v ListContainerRegistryCredentialIdsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ListContainerRegistryCredentialIdsValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	for _, element := range req.ConfigValue.Elements() {
		value := element.(types.String)
		if value.IsUnknown() {
			continue
		}

		if value.IsNull() || !containerRegistryCredentialIDPattern.MatchString(value.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid container registry credential ID",
				"`container_reg_ids` must contain Seqera Platform credential IDs encoded as 1 to 22 alphanumeric characters, such as `seqera_container_registry_credential.acr.id`.",
			)
			return
		}
	}
}

// ContainerRegistryCredentialIdsValidator validates container_reg_ids values.
func ContainerRegistryCredentialIdsValidator() validator.List {
	return ListContainerRegistryCredentialIdsValidator{}
}
