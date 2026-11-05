package objectvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// makeBackendStrategyRequest builds a request for a `config` object holding
// intelligent_compute_config.backend_strategy. schedRaw is the raw value of
// the intelligent_compute_config block.
func makeBackendStrategyRequest(schedRaw func(tftypes.Object) tftypes.Value) validator.ObjectRequest {
	schedType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"backend_strategy": tftypes.String}}
	configType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"intelligent_compute_config": schedType}}
	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"config": configType}}

	configRaw := tftypes.NewValue(configType, map[string]tftypes.Value{"intelligent_compute_config": schedRaw(schedType)})
	rootRaw := tftypes.NewValue(rootType, map[string]tftypes.Value{"config": configRaw})

	schema := resourceschema.Schema{Attributes: map[string]resourceschema.Attribute{
		"config": resourceschema.SingleNestedAttribute{Optional: true, Attributes: map[string]resourceschema.Attribute{
			"intelligent_compute_config": resourceschema.SingleNestedAttribute{Optional: true, Attributes: map[string]resourceschema.Attribute{
				"backend_strategy": resourceschema.StringAttribute{Optional: true, Computed: true},
			}},
		}},
	}}

	// The validator only checks ConfigValue for null/unknown, then reads from Config.
	configValue, _ := types.ObjectValue(map[string]attr.Type{}, map[string]attr.Value{})
	return validator.ObjectRequest{
		Path:        path.Root("config"),
		ConfigValue: configValue,
		Config:      tfsdk.Config{Schema: schema, Raw: rootRaw},
	}
}

func TestBackendStrategyVMOnlyValidator(t *testing.T) {
	strategy := func(v interface{}) func(tftypes.Object) tftypes.Value {
		return func(st tftypes.Object) tftypes.Value {
			return tftypes.NewValue(st, map[string]tftypes.Value{"backend_strategy": tftypes.NewValue(tftypes.String, v)})
		}
	}
	tests := []struct {
		name        string
		sched       func(tftypes.Object) tftypes.Value
		expectError bool
	}{
		{"no intelligent_compute_config passes", func(st tftypes.Object) tftypes.Value { return tftypes.NewValue(st, nil) }, false},
		{"unset backend_strategy passes", strategy(nil), false},
		{"unknown backend_strategy passes", strategy(tftypes.UnknownValue), false},
		{"VM passes", strategy("VM"), false},
		{"deprecated ECS is rejected", strategy("ECS"), true},
		{"deprecated EC2 is rejected", strategy("EC2"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &validator.ObjectResponse{}
			BackendStrategyVMOnlyValidator().ValidateObject(context.Background(), makeBackendStrategyRequest(tt.sched), resp)

			if tt.expectError != resp.Diagnostics.HasError() {
				t.Fatalf("expectError=%v, got diagnostics: %v", tt.expectError, resp.Diagnostics)
			}
			if !tt.expectError {
				return
			}
			want := path.Root("config").AtName("intelligent_compute_config").AtName("backend_strategy")
			if got := resp.Diagnostics.Errors()[0].(interface{ Path() path.Path }).Path(); !got.Equal(want) {
				t.Errorf("error path = %s, want %s", got, want)
			}
		})
	}
}
