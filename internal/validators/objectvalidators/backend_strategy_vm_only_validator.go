package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = ObjectBackendStrategyVMOnlyValidator{}

// ObjectBackendStrategyVMOnlyValidator enforces that
// intelligent_compute_config.backend_strategy is "VM" (or unset). VM is the
// only backend the API accepts when creating a compute environment: the
// AWS-only ECS and EC2 backends are deprecated, and Azure and Google never
// supported them. Any other value is rejected at plan time instead of failing
// on apply.
type ObjectBackendStrategyVMOnlyValidator struct{}

func (v ObjectBackendStrategyVMOnlyValidator) Description(_ context.Context) string {
	return "Validates that intelligent_compute_config.backend_strategy is \"VM\" (or unset)."
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
			"Intelligent Compute only supports the `VM` backend. The `ECS` and `EC2` backends are deprecated. "+
				"Set `backend_strategy = \"VM\"` or omit it. Omitting it does not replace an existing compute environment.",
		)
	}
}

// BackendStrategyVMOnlyValidator returns a validator that restricts
// intelligent_compute_config.backend_strategy to "VM" (or unset).
func BackendStrategyVMOnlyValidator() validator.Object {
	return ObjectBackendStrategyVMOnlyValidator{}
}
