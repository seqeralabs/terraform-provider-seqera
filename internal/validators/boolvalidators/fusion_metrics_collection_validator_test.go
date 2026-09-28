package boolvalidators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// makeFusionMetricsRequest builds a validator.BoolRequest shaped like a Batch
// compute environment: `fusion_metrics_collection_enabled` at the root, with
// `enable_fusion` one level down inside the `config` object. A nil enableFusion
// means the attribute is null; unknownFusion overrides it with an unknown value,
// and nullConfig makes the whole `config` object null.
func makeFusionMetricsRequest(metrics types.Bool, enableFusion *bool, unknownFusion, nullConfig bool) validator.BoolRequest {
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"enable_fusion": tftypes.Bool,
		},
	}
	rootType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"fusion_metrics_collection_enabled": tftypes.Bool,
			"config":                            configType,
		},
	}

	var metricsRaw tftypes.Value
	switch {
	case metrics.IsUnknown():
		metricsRaw = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	case metrics.IsNull():
		metricsRaw = tftypes.NewValue(tftypes.Bool, nil)
	default:
		metricsRaw = tftypes.NewValue(tftypes.Bool, metrics.ValueBool())
	}

	var fusionRaw tftypes.Value
	switch {
	case unknownFusion:
		fusionRaw = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	case enableFusion == nil:
		fusionRaw = tftypes.NewValue(tftypes.Bool, nil)
	default:
		fusionRaw = tftypes.NewValue(tftypes.Bool, *enableFusion)
	}

	configRaw := tftypes.NewValue(configType, map[string]tftypes.Value{
		"enable_fusion": fusionRaw,
	})
	if nullConfig {
		configRaw = tftypes.NewValue(configType, nil)
	}

	return validator.BoolRequest{
		Path:        path.Root("fusion_metrics_collection_enabled"),
		ConfigValue: metrics,
		Config: tfsdk.Config{
			Schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"fusion_metrics_collection_enabled": resourceschema.BoolAttribute{Optional: true},
					"config": resourceschema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]resourceschema.Attribute{
							"enable_fusion": resourceschema.BoolAttribute{Optional: true},
						},
					},
				},
			},
			Raw: tftypes.NewValue(rootType, map[string]tftypes.Value{
				"fusion_metrics_collection_enabled": metricsRaw,
				"config":                            configRaw,
			}),
		},
	}
}

func TestFusionMetricsCollectionValidator(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name          string
		metrics       types.Bool
		enableFusion  *bool
		unknownFusion bool
		nullConfig    bool
		expectError   bool
	}{
		{
			name:         "true with fusion enabled passes",
			metrics:      types.BoolValue(true),
			enableFusion: boolPtr(true),
		},
		{
			// The backend rejects this combination with a 400:
			// "Fusion metrics collection cannot be enabled on a compute
			// environment without Fusion enabled".
			name:         "true with fusion disabled errors",
			metrics:      types.BoolValue(true),
			enableFusion: boolPtr(false),
			expectError:  true,
		},
		{
			name:         "true with fusion unset errors",
			metrics:      types.BoolValue(true),
			enableFusion: nil,
			expectError:  true,
		},
		{
			name:        "true with null config errors",
			metrics:     types.BoolValue(true),
			nullConfig:  true,
			expectError: true,
		},
		{
			name:         "false with fusion disabled passes",
			metrics:      types.BoolValue(false),
			enableFusion: boolPtr(false),
		},
		{
			name:         "false with fusion unset passes",
			metrics:      types.BoolValue(false),
			enableFusion: nil,
		},
		{
			name:         "null metrics passes",
			metrics:      types.BoolNull(),
			enableFusion: nil,
		},
		{
			name:         "unknown metrics is skipped",
			metrics:      types.BoolUnknown(),
			enableFusion: nil,
		},
		{
			// enable_fusion interpolated from elsewhere is not known at plan time;
			// erroring here would be a false positive, so defer to the API.
			name:          "true with unknown enable_fusion is skipped",
			metrics:       types.BoolValue(true),
			unknownFusion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := makeFusionMetricsRequest(tt.metrics, tt.enableFusion, tt.unknownFusion, tt.nullConfig)
			resp := &validator.BoolResponse{}

			FusionMetricsCollectionValidator().ValidateBool(context.Background(), req, resp)

			if got := resp.Diagnostics.HasError(); got != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v: %v", tt.expectError, got, resp.Diagnostics)
			}

			if tt.expectError {
				detail := resp.Diagnostics.Errors()[0].Detail()
				if !strings.Contains(detail, "'config.enable_fusion' must also be set to true") {
					t.Errorf("error detail should explain the Fusion requirement, got: %q", detail)
				}
			}
		})
	}
}
