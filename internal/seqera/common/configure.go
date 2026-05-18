package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
)

// ConfigureClient extracts the Seqera SDK client from a Terraform resource
// or data source Configure request's provider data. Returns nil when
// providerData is unset (first plan/apply pass before the provider is
// configured) or the wrong type (diagnostic appended). Callers must
// nil-check before assigning — unconditional assignment would clobber a
// previously-set client on subsequent Configure invocations.
func ConfigureClient(providerData any) (*sdk.Seqera, diag.Diagnostics) {
	var diags diag.Diagnostics
	if providerData == nil {
		return nil, diags
	}
	client, ok := providerData.(*sdk.Seqera)
	if !ok {
		diags.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *sdk.Seqera, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil, diags
	}
	return client, diags
}
