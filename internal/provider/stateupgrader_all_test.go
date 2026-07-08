package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestAllRegisteredUpgradersDropUnknownAttributes drives EVERY registered state
// upgrader of EVERY resource through the exact check the framework performs: run
// the upgrader, then strictly decode its returned DynamicValue against the
// current schema (no IgnoreUndefinedAttributes). Before the default
// lenient-decode pattern, an upgrader that returned prior state still carrying an
// attribute since removed from the schema failed here with "unsupported
// attribute" (issue #228).
//
// The fixture injects unknown attributes at the top level AND inside every
// top-level object attribute, standing in for attributes removed from the schema
// since the prior state was written. A correct upgrader drops them all (at any
// depth) so the state decodes cleanly. Enumerating resources via Resources() and
// their upgraders via UpgradeState() keeps this exhaustive as resources are added.
func TestAllRegisteredUpgradersDropUnknownAttributes(t *testing.T) {
	ctx := context.Background()

	p := &SeqeraProvider{}
	var providerMeta provider.MetadataResponse
	p.Metadata(ctx, provider.MetadataRequest{}, &providerMeta)

	upgradersSeen := 0

	for _, newResource := range p.Resources(ctx) {
		res := newResource()

		upgradeable, ok := res.(resource.ResourceWithUpgradeState)
		if !ok {
			continue
		}

		var meta resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: providerMeta.TypeName}, &meta)

		var schemaResp resource.SchemaResponse
		res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			t.Fatalf("%s: schema errors: %v", meta.TypeName, schemaResp.Diagnostics)
		}
		schemaType := schemaResp.Schema.Type().TerraformType(ctx)

		// Build a prior-state fixture full of attributes the current schema does
		// not define — one at the top level and one inside each top-level object
		// attribute — to exercise recursive dropping.
		fixture := map[string]interface{}{
			"__removed_top_level_attribute": "should be dropped by the upgrade",
		}
		if obj, ok := schemaType.(tftypes.Object); ok {
			for name, attrType := range obj.AttributeTypes {
				if _, isObject := attrType.(tftypes.Object); isObject {
					fixture[name] = map[string]interface{}{
						"__removed_nested_attribute": "should be dropped by the upgrade",
					}
				}
			}
		}

		priorJSON, err := json.Marshal(fixture)
		if err != nil {
			t.Fatalf("%s: marshal fixture: %v", meta.TypeName, err)
		}

		for version, upgrader := range upgradeable.UpgradeState(ctx) {
			upgradersSeen++
			name := fmt.Sprintf("%s/v%d", meta.TypeName, version)
			t.Run(name, func(t *testing.T) {
				if upgrader.StateUpgrader == nil {
					t.Fatalf("no StateUpgrader function registered")
				}

				req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: priorJSON}}
				resp := &resource.UpgradeStateResponse{}
				upgrader.StateUpgrader(ctx, req, resp)

				if resp.Diagnostics.HasError() {
					t.Fatalf("upgrader diagnostics: %v", resp.Diagnostics)
				}
				if resp.DynamicValue == nil {
					t.Fatalf("upgrader returned no DynamicValue")
				}

				// The exact strict decode the framework performs on an upgrader's
				// returned DynamicValue. Any undropped unknown attribute fails here.
				if _, err := resp.DynamicValue.Unmarshal(schemaType); err != nil {
					t.Fatalf("upgraded state failed to decode against current schema: %v", err)
				}
			})
		}
	}

	if upgradersSeen == 0 {
		t.Fatal("no registered upgraders were exercised; enumeration is broken")
	}
	t.Logf("exercised %d registered upgraders", upgradersSeen)
}
