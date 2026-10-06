package boolvalidators

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = privateAddressRequiresSubnetValidator{}

// privateAddressRequiresSubnetValidator enforces that the sibling 'subnet_id' is set
// whenever 'use_private_address' is configured on an Azure Batch compute environment.
// Unlike RequiresSiblingStringSet this is presence-based: an explicit false is also
// rejected, since the API only accepts the field together with a dedicated subnet.
type privateAddressRequiresSubnetValidator struct{}

func (v privateAddressRequiresSubnetValidator) Description(_ context.Context) string {
	return "validates that subnet_id is set when this attribute is configured"
}

func (v privateAddressRequiresSubnetValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v privateAddressRequiresSubnetValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var subnetID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("subnet_id"), &subnetID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase (interpolated from another resource).
	if subnetID.IsUnknown() {
		return
	}

	if subnetID.IsNull() || strings.TrimSpace(subnetID.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Subnet Required for Private Addressing",
			"'use_private_address' can only be set when 'subnet_id' is set to a dedicated Azure VNet subnet. "+
				"The subnet must have outbound connectivity (e.g. a NAT gateway) for pool nodes without public IPs.",
		)
	}
}

// PrivateAddressRequiresSubnetValidator enforces that 'subnet_id' is set whenever
// 'use_private_address' is configured.
func PrivateAddressRequiresSubnetValidator() validator.Bool {
	return privateAddressRequiresSubnetValidator{}
}
