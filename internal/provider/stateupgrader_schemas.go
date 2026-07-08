package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	stateupgraders "github.com/seqeralabs/terraform-provider-seqera/internal/stateupgraders"
)

// The state upgraders live in the stateupgraders package because the
// Speakeasy-generated UpgradeState() registrations reference them by name. To
// drop attributes removed from a schema since a prior state was written, those
// upgraders need the current resource schema. The generated registrations cannot
// set StateUpgrader.PriorSchema, and stateupgraders cannot import this package
// (import cycle), so we register every upgradeable resource's current schema
// type here at package initialization. See docs-internal/STATE_UPGRADER_GUIDE.md.
func init() {
	ctx := context.Background()

	p := &SeqeraProvider{}
	var providerMeta provider.MetadataResponse
	p.Metadata(ctx, provider.MetadataRequest{}, &providerMeta)

	for _, newResource := range p.Resources(ctx) {
		res := newResource()

		// Only resources with state upgraders need their schema registered.
		if _, ok := res.(resource.ResourceWithUpgradeState); !ok {
			continue
		}

		var metaResp resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: providerMeta.TypeName}, &metaResp)

		var schemaResp resource.SchemaResponse
		res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			continue
		}

		stateupgraders.RegisterSchemaType(metaResp.TypeName, schemaResp.Schema.Type().TerraformType(ctx))
	}
}
