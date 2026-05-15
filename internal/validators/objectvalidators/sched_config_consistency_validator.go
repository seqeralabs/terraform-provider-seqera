package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = ObjectSchedConfigConsistencyValidatorValidator{}

type ObjectSchedConfigConsistencyValidatorValidator struct{}

func (v ObjectSchedConfigConsistencyValidatorValidator) Description(_ context.Context) string {
	return "Validates that intelligent_compute_config is omitted when intelligent_compute_enabled is false. The config is optional when enabled is true — the API uses defaults if it's null."
}

func (v ObjectSchedConfigConsistencyValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ObjectSchedConfigConsistencyValidatorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var schedEnabled types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("intelligent_compute_enabled"), &schedEnabled)...)
	if resp.Diagnostics.HasError() || schedEnabled.IsUnknown() {
		return
	}

	var schedConfig types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("intelligent_compute_config"), &schedConfig)...)
	if resp.Diagnostics.HasError() || schedConfig.IsUnknown() {
		return
	}

	enabled := !schedEnabled.IsNull() && schedEnabled.ValueBool()
	configSet := !schedConfig.IsNull()

	if !enabled && configSet {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("intelligent_compute_config"),
			"Unexpected intelligent_compute_config",
			"`intelligent_compute_config` must be omitted when `intelligent_compute_enabled = false` (Classic mode). "+
				"Remove the `intelligent_compute_config` block, or set `intelligent_compute_enabled = true` to enable Seqera Intelligent Compute.",
		)
	}
}

func SchedConfigConsistencyValidator() validator.Object {
	return ObjectSchedConfigConsistencyValidatorValidator{}
}
