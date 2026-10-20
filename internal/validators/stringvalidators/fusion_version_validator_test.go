package stringvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestFusionVersionValidator(t *testing.T) {
	t.Parallel()

	sch := schema.Schema{Attributes: map[string]schema.Attribute{
		"fusion_version":                   schema.StringAttribute{Optional: true},
		"override_nextflow_fusion_version": schema.BoolAttribute{Optional: true},
	}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"fusion_version":                   tftypes.String,
		"override_nextflow_fusion_version": tftypes.Bool,
	}}

	tests := map[string]struct {
		version   types.String
		override  tftypes.Value
		wantError bool
	}{
		"version null":                  {types.StringNull(), tftypes.NewValue(tftypes.Bool, nil), false},
		"version unknown":               {types.StringUnknown(), tftypes.NewValue(tftypes.Bool, nil), false},
		"version set, override true":    {types.StringValue("2.4.0"), tftypes.NewValue(tftypes.Bool, true), false},
		"version set, override false":   {types.StringValue("2.4.0"), tftypes.NewValue(tftypes.Bool, false), true},
		"version set, override null":    {types.StringValue("2.4.0"), tftypes.NewValue(tftypes.Bool, nil), true},
		"version set, override unknown": {types.StringValue("2.4.0"), tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var versionVal tftypes.Value
			switch {
			case tc.version.IsNull():
				versionVal = tftypes.NewValue(tftypes.String, nil)
			case tc.version.IsUnknown():
				versionVal = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			default:
				versionVal = tftypes.NewValue(tftypes.String, tc.version.ValueString())
			}

			req := validator.StringRequest{
				Path:        path.Root("fusion_version"),
				ConfigValue: tc.version,
				Config: tfsdk.Config{
					Schema: sch,
					Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
						"fusion_version":                   versionVal,
						"override_nextflow_fusion_version": tc.override,
					}),
				},
			}
			resp := &validator.StringResponse{}
			FusionVersionValidator().ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("wantError=%v, diagnostics=%v", tc.wantError, resp.Diagnostics)
			}
		})
	}
}
