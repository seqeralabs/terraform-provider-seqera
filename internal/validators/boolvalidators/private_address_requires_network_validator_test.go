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

// makePrivateAddressRequest builds a validator.BoolRequest for use_private_address
// alongside a sibling network attribute. A nil network means the attribute is null;
// unknownNetwork overrides it with an unknown value.
func makePrivateAddressRequest(usePrivateAddress types.Bool, network *string, unknownNetwork bool) validator.BoolRequest {
	ct := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"use_private_address": tftypes.Bool,
			"network":             tftypes.String,
		},
	}

	var boolRaw tftypes.Value
	switch {
	case usePrivateAddress.IsUnknown():
		boolRaw = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	case usePrivateAddress.IsNull():
		boolRaw = tftypes.NewValue(tftypes.Bool, nil)
	default:
		boolRaw = tftypes.NewValue(tftypes.Bool, usePrivateAddress.ValueBool())
	}

	var networkRaw tftypes.Value
	switch {
	case unknownNetwork:
		networkRaw = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	case network == nil:
		networkRaw = tftypes.NewValue(tftypes.String, nil)
	default:
		networkRaw = tftypes.NewValue(tftypes.String, *network)
	}

	return validator.BoolRequest{
		Path:        path.Root("use_private_address"),
		ConfigValue: usePrivateAddress,
		Config: tfsdk.Config{
			Schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"use_private_address": resourceschema.BoolAttribute{Optional: true},
					"network":             resourceschema.StringAttribute{Optional: true},
				},
			},
			Raw: tftypes.NewValue(ct, map[string]tftypes.Value{
				"use_private_address": boolRaw,
				"network":             networkRaw,
			}),
		},
	}
}

func TestPrivateAddressRequiresNetworkValidator(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name              string
		usePrivateAddress types.Bool
		network           *string
		unknownNetwork    bool
		expectError       bool
	}{
		{
			name:              "true with network set passes",
			usePrivateAddress: types.BoolValue(true),
			network:           str("projects/p/global/networks/vpc-main"),
		},
		{
			name:              "true with short network name passes",
			usePrivateAddress: types.BoolValue(true),
			network:           str("vpc-main"),
		},
		{
			name:              "true without network errors",
			usePrivateAddress: types.BoolValue(true),
			network:           nil,
			expectError:       true,
		},
		{
			name:              "true with empty network errors",
			usePrivateAddress: types.BoolValue(true),
			network:           str(""),
			expectError:       true,
		},
		{
			name:              "true with whitespace-only network errors",
			usePrivateAddress: types.BoolValue(true),
			network:           str("   "),
			expectError:       true,
		},
		{
			// The key difference from the native x-speakeasy-required-with /
			// AlsoRequires behaviour, which is presence-based and would error here.
			name:              "false without network passes",
			usePrivateAddress: types.BoolValue(false),
			network:           nil,
		},
		{
			name:              "null without network passes",
			usePrivateAddress: types.BoolNull(),
			network:           nil,
		},
		{
			name:              "unknown use_private_address is skipped",
			usePrivateAddress: types.BoolUnknown(),
			network:           nil,
		},
		{
			// network interpolated from another resource is not yet known at plan time.
			name:              "true with unknown network is skipped",
			usePrivateAddress: types.BoolValue(true),
			unknownNetwork:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := makePrivateAddressRequest(tt.usePrivateAddress, tt.network, tt.unknownNetwork)
			resp := &validator.BoolResponse{}

			PrivateAddressRequiresNetworkValidator().ValidateBool(context.Background(), req, resp)

			if got := resp.Diagnostics.HasError(); got != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v: %v", tt.expectError, got, resp.Diagnostics)
			}

			if tt.expectError {
				detail := resp.Diagnostics.Errors()[0].Detail()
				if !strings.Contains(detail, "'network' must be set") {
					t.Errorf("error detail should explain the network requirement, got: %q", detail)
				}
			}
		})
	}
}
