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

func TestLineageManualFieldValidator(t *testing.T) {
	t.Parallel()

	sch := schema.Schema{Attributes: map[string]schema.Attribute{
		"provisioning_mode": schema.StringAttribute{Optional: true},
		"store_name":        schema.StringAttribute{Optional: true},
		"topic_arn":         schema.StringAttribute{Optional: true},
	}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"provisioning_mode": tftypes.String,
		"store_name":        tftypes.String,
		"topic_arn":         tftypes.String,
	}}

	null := tftypes.NewValue(tftypes.String, nil)
	unknown := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }

	tests := map[string]struct {
		attr      string
		mode      tftypes.Value
		field     tftypes.Value
		wantError bool
	}{
		"MANUAL, field null":       {"store_name", str("MANUAL"), null, true},
		"MANUAL, field set":        {"store_name", str("MANUAL"), str("bucket"), false},
		"AUTO, field set":          {"store_name", str("AUTO"), str("bucket"), true},
		"AUTO, field null":         {"store_name", str("AUTO"), null, false},
		"mode null, field set":     {"store_name", null, str("bucket"), true},
		"mode null, field null":    {"store_name", null, null, false},
		"mode unknown, field set":  {"store_name", unknown, str("bucket"), false},
		"mode unknown, field null": {"store_name", unknown, null, false},
		"MANUAL, field unknown":    {"store_name", str("MANUAL"), unknown, false},
		"AUTO, field unknown":      {"store_name", str("AUTO"), unknown, false},
		"topic_arn MANUAL, null":   {"topic_arn", str("MANUAL"), null, true},
		"topic_arn AUTO, set":      {"topic_arn", str("AUTO"), str("arn:aws:sns:x"), true},
		"topic_arn MANUAL, set":    {"topic_arn", str("MANUAL"), str("arn:aws:sns:x"), false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vals := map[string]tftypes.Value{
				"provisioning_mode": tc.mode,
				"store_name":        null,
				"topic_arn":         null,
			}
			vals[tc.attr] = tc.field

			var cv types.String
			switch {
			case tc.field.IsNull():
				cv = types.StringNull()
			case !tc.field.IsKnown():
				cv = types.StringUnknown()
			default:
				var s string
				_ = tc.field.As(&s)
				cv = types.StringValue(s)
			}

			req := validator.StringRequest{
				Path:        path.Root(tc.attr),
				ConfigValue: cv,
				Config: tfsdk.Config{
					Schema: sch,
					Raw:    tftypes.NewValue(objType, vals),
				},
			}
			resp := &validator.StringResponse{}
			LineageManualFieldValidator().ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("wantError=%v, diagnostics=%v", tc.wantError, resp.Diagnostics)
			}
		})
	}
}
