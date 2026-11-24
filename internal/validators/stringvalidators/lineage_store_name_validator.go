package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringLineageStoreNameValidator{}

// StringLineageStoreNameValidator validates the deprecated store_name on
// seqera_workspace_lineage_settings: not allowed in AUTO mode. In MANUAL mode it
// is optional; the requirement for a store and the conflict with store_location
// are reported by LineageStoreLocationValidator. See lineage_store_common.go.
type StringLineageStoreNameValidator struct{}

func (v StringLineageStoreNameValidator) Description(_ context.Context) string {
	return "must not be set when provisioning_mode is AUTO"
}

func (v StringLineageStoreNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringLineageStoreNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	lineageStoreAutoCheck(ctx, req, resp)
}

func LineageStoreNameValidator() validator.String {
	return StringLineageStoreNameValidator{}
}
