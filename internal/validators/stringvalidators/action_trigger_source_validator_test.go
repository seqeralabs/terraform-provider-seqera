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

func TestActionTriggerSourceValidator(t *testing.T) {
	ctx := context.Background()

	blockAttr := schema.SingleNestedAttribute{
		Optional:   true,
		Attributes: map[string]schema.Attribute{"x": schema.StringAttribute{Optional: true}},
	}
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"source":          schema.StringAttribute{Required: true},
		"bucket":          blockAttr,
		"cron":            blockAttr,
		"pipeline_status": blockAttr,
	}}
	objType := s.Type().TerraformType(ctx).(tftypes.Object)
	blockType := objType.AttributeTypes["bucket"]

	set := tftypes.NewValue(blockType, map[string]tftypes.Value{"x": tftypes.NewValue(tftypes.String, "v")})
	null := tftypes.NewValue(blockType, nil)
	unknown := tftypes.NewValue(blockType, tftypes.UnknownValue)

	tests := []struct {
		name       string
		source     string // "" means unset
		blocks     map[string]tftypes.Value
		wantErrors int
	}{
		{"unset source takes no block", "", nil, 0},
		{"unset source with a block", "", map[string]tftypes.Value{"cron": set}, 1},
		{"cron with cron block", "cron", map[string]tftypes.Value{"cron": set}, 0},
		{"bucket with bucket block", "bucket", map[string]tftypes.Value{"bucket": set}, 0},
		{"pipeline_status with its block", "pipeline_status", map[string]tftypes.Value{"pipeline_status": set}, 0},
		{"tower takes no block", "tower", nil, 0},
		{"github takes no block", "github", nil, 0},
		{"missing matching block", "cron", nil, 1},
		{"wrong block only", "cron", map[string]tftypes.Value{"bucket": set}, 2},
		{"extra block", "cron", map[string]tftypes.Value{"cron": set, "pipeline_status": set}, 1},
		{"tower with a block", "tower", map[string]tftypes.Value{"cron": set}, 1},
		{"unknown block is not judged", "cron", map[string]tftypes.Value{"cron": unknown, "bucket": unknown}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceValue, configValue := tftypes.NewValue(tftypes.String, tt.source), types.StringValue(tt.source)
			if tt.source == "" {
				sourceValue, configValue = tftypes.NewValue(tftypes.String, nil), types.StringNull()
			}
			attrs := map[string]tftypes.Value{
				"source":          sourceValue,
				"bucket":          null,
				"cron":            null,
				"pipeline_status": null,
			}
			for k, v := range tt.blocks {
				attrs[k] = v
			}
			config := tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, attrs)}

			resp := &validator.StringResponse{}
			ActionTriggerSourceValidator().ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("source"),
				ConfigValue: configValue,
				Config:      config,
			}, resp)

			if got := resp.Diagnostics.ErrorsCount(); got != tt.wantErrors {
				t.Errorf("got %d errors, want %d: %v", got, tt.wantErrors, resp.Diagnostics)
			}
		})
	}
}
