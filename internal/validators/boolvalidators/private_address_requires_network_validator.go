package boolvalidators

import (
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// PrivateAddressRequiresNetworkValidator enforces that 'network' is set when
// 'use_private_address' is true. Instances launched without an external IP cannot
// fall back to the project's 'default' network, so the API rejects the combination.
func PrivateAddressRequiresNetworkValidator() validator.Bool {
	return RequiresSiblingStringSet(
		"network",
		"Network Required for Private Addressing",
		"When 'use_private_address' is true, 'network' must be set to an explicit VPC network. "+
			"Instances without an external IP cannot use the project's 'default' network; "+
			"the network must also have Cloud NAT and Private Google Access configured on the subnetwork.",
	)
}
