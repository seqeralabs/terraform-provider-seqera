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

// makePrivateAddressSubnetRequest builds a validator.BoolRequest for use_private_address
// alongside a sibling subnet_id attribute. A nil subnetID means the attribute is null;
// unknownSubnet overrides it with an unknown value.
func makePrivateAddressSubnetRequest(usePrivateAddress types.Bool, subnetID *string, unknownSubnet bool) validator.BoolRequest {
	ct := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"use_private_address": tftypes.Bool,
			"subnet_id":           tftypes.String,
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

	var subnetRaw tftypes.Value
	switch {
	case unknownSubnet:
		subnetRaw = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	case subnetID == nil:
		subnetRaw = tftypes.NewValue(tftypes.String, nil)
	default:
		subnetRaw = tftypes.NewValue(tftypes.String, *subnetID)
	}

	return validator.BoolRequest{
		Path:        path.Root("use_private_address"),
		ConfigValue: usePrivateAddress,
		Config: tfsdk.Config{
			Schema: resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"use_private_address": resourceschema.BoolAttribute{Optional: true},
					"subnet_id":           resourceschema.StringAttribute{Optional: true},
				},
			},
			Raw: tftypes.NewValue(ct, map[string]tftypes.Value{
				"use_private_address": boolRaw,
				"subnet_id":           subnetRaw,
			}),
		},
	}
}

func TestPrivateAddressRequiresSubnetValidator(t *testing.T) {
	str := func(s string) *string { return &s }
	subnet := "/subscriptions/0000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/snet"

	tests := []struct {
		name              string
		usePrivateAddress types.Bool
		subnetID          *string
		unknownSubnet     bool
		expectError       bool
	}{
		{
			name:              "true with subnet set passes",
			usePrivateAddress: types.BoolValue(true),
			subnetID:          str(subnet),
		},
		{
			name:              "false with subnet set passes",
			usePrivateAddress: types.BoolValue(false),
			subnetID:          str(subnet),
		},
		{
			name:              "true without subnet errors",
			usePrivateAddress: types.BoolValue(true),
			expectError:       true,
		},
		{
			// Presence-based: the API rejects the field without a subnet regardless of value.
			name:              "false without subnet errors",
			usePrivateAddress: types.BoolValue(false),
			expectError:       true,
		},
		{
			name:              "true with whitespace-only subnet errors",
			usePrivateAddress: types.BoolValue(true),
			subnetID:          str("   "),
			expectError:       true,
		},
		{
			name:              "null without subnet passes",
			usePrivateAddress: types.BoolNull(),
		},
		{
			name:              "unknown use_private_address is skipped",
			usePrivateAddress: types.BoolUnknown(),
		},
		{
			// subnet_id interpolated from another resource is not yet known at plan time.
			name:              "true with unknown subnet is skipped",
			usePrivateAddress: types.BoolValue(true),
			unknownSubnet:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := makePrivateAddressSubnetRequest(tt.usePrivateAddress, tt.subnetID, tt.unknownSubnet)
			resp := &validator.BoolResponse{}

			PrivateAddressRequiresSubnetValidator().ValidateBool(context.Background(), req, resp)

			if got := resp.Diagnostics.HasError(); got != tt.expectError {
				t.Fatalf("expected error=%v, got error=%v: %v", tt.expectError, got, resp.Diagnostics)
			}

			if tt.expectError {
				detail := resp.Diagnostics.Errors()[0].Detail()
				if !strings.Contains(detail, "'subnet_id' is set") {
					t.Errorf("error detail should explain the subnet requirement, got: %q", detail)
				}
			}
		})
	}
}
