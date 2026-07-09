package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = ObjectBackendStrategyVMOnlyValidator{}

// ObjectBackendStrategyVMOnlyValidator enforces that
// intelligent_compute_config.backend_strategy is "VM" (or unset) on cloud
// platforms where the backend only supports the VM backend. On Azure and
// Google the backend ignores backend_strategy and hardcodes VM (the ECS/EC2
// backends are AWS-only), so any other value is rejected at plan time rather
// than being silently ignored.
type ObjectBackendStrategyVMOnlyValidator struct{}

func (v ObjectBackendStrategyVMOnlyValidator) Description(_ context.Context) string {
	return "Validates that intelligent_compute_config.backend_strategy is \"VM\" (or unset) on platforms that only support the VM backend."
}

func (v ObjectBackendStrategyVMOnlyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ObjectBackendStrategyVMOnlyValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var backendStrategy types.String
	strategyPath := req.Path.AtName("intelligent_compute_config").AtName("backend_strategy")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, strategyPath, &backendStrategy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unset (or the block absent) is fine — the backend defaults to VM.
	if backendStrategy.IsNull() || backendStrategy.IsUnknown() || backendStrategy.ValueString() == "" {
		return
	}

	if backendStrategy.ValueString() != "VM" {
		resp.Diagnostics.AddAttributeError(
			strategyPath,
			"Unsupported backend_strategy",
			"This compute environment only supports the `VM` backend for Intelligent Compute. "+
				"Set `backend_strategy = \"VM\"` or omit it (the ECS/EC2 backends are AWS-only).",
		)
	}
}

// BackendStrategyVMOnlyValidator returns a validator that restricts
// intelligent_compute_config.backend_strategy to "VM" (or unset).
func BackendStrategyVMOnlyValidator() validator.Object {
	return ObjectBackendStrategyVMOnlyValidator{}
}
