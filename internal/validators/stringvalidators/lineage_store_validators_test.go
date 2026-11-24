package stringvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestLineageStoreValidators(t *testing.T) {
	t.Parallel()

	sch := schema.Schema{Attributes: map[string]schema.Attribute{
		"provisioning_mode": schema.StringAttribute{Optional: true},
		"store_location":    schema.StringAttribute{Optional: true},
		"store_name":        schema.StringAttribute{Optional: true},
	}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"provisioning_mode": tftypes.String,
		"store_location":    tftypes.String,
		"store_name":        tftypes.String,
	}}

	null := tftypes.NewValue(tftypes.String, nil)
	unknown := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	loc := str("s3://bucket/lineage")
	name := str("bucket")

	toTypes := func(v tftypes.Value) types.String {
		switch {
		case v.IsNull():
			return types.StringNull()
		case !v.IsKnown():
			return types.StringUnknown()
		}
		var s string
		_ = v.As(&s)
		return types.StringValue(s)
	}

	tests := map[string]struct {
		attr      string // attribute under validation
		mode      tftypes.Value
		location  tftypes.Value
		storeName tftypes.Value
		wantError bool
	}{
		// store_location
		"location: MANUAL, location set":                {"store_location", str("MANUAL"), loc, null, false},
		"location: MANUAL, only deprecated name set":    {"store_location", str("MANUAL"), null, name, false},
		"location: MANUAL, neither set":                 {"store_location", str("MANUAL"), null, null, true},
		"location: MANUAL, both set":                    {"store_location", str("MANUAL"), loc, name, true},
		"location: MANUAL, location unknown, name set":  {"store_location", str("MANUAL"), unknown, name, false},
		"location: MANUAL, location null, name unknown": {"store_location", str("MANUAL"), null, unknown, false},
		"location: AUTO, location set":                  {"store_location", str("AUTO"), loc, null, true},
		"location: AUTO, neither set":                   {"store_location", str("AUTO"), null, null, false},
		"location: mode null, location set":             {"store_location", null, loc, null, true},
		"location: mode null, neither set":              {"store_location", null, null, null, false},
		"location: mode unknown, location set":          {"store_location", unknown, loc, null, false},
		"location: mode unknown, neither set":           {"store_location", unknown, null, null, false},
		// deprecated store_name
		"name: MANUAL, name set":    {"store_name", str("MANUAL"), null, name, false},
		"name: MANUAL, name null":   {"store_name", str("MANUAL"), loc, null, false},
		"name: AUTO, name set":      {"store_name", str("AUTO"), null, name, true},
		"name: AUTO, name null":     {"store_name", str("AUTO"), null, null, false},
		"name: AUTO, name unknown":  {"store_name", str("AUTO"), null, unknown, false},
		"name: mode null, name set": {"store_name", null, null, name, true},
		"name: mode unknown, set":   {"store_name", unknown, null, name, false},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()

			vals := map[string]tftypes.Value{
				"provisioning_mode": tc.mode,
				"store_location":    tc.location,
				"store_name":        tc.storeName,
			}

			v := LineageStoreLocationValidator()
			if tc.attr == "store_name" {
				v = LineageStoreNameValidator()
			}

			req := validator.StringRequest{
				Path:        path.Root(tc.attr),
				ConfigValue: toTypes(vals[tc.attr]),
				Config: tfsdk.Config{
					Schema: sch,
					Raw:    tftypes.NewValue(objType, vals),
				},
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("wantError=%v, diagnostics=%v", tc.wantError, resp.Diagnostics)
			}
		})
	}
}
