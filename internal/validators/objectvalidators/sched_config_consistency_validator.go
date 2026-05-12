package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = ObjectSchedConfigConsistencyValidatorValidator{}

type ObjectSchedConfigConsistencyValidatorValidator struct{}

func (v ObjectSchedConfigConsistencyValidatorValidator) Description(_ context.Context) string {
	return "Validates that sched_config is set when sched_enabled is true, and is omitted when sched_enabled is false."
}

func (v ObjectSchedConfigConsistencyValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ObjectSchedConfigConsistencyValidatorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var schedEnabled types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("sched_enabled"), &schedEnabled)...)
	if resp.Diagnostics.HasError() || schedEnabled.IsUnknown() {
		return
	}

	var schedConfig types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.AtName("sched_config"), &schedConfig)...)
	if resp.Diagnostics.HasError() || schedConfig.IsUnknown() {
		return
	}

	enabled := !schedEnabled.IsNull() && schedEnabled.ValueBool()
	configSet := !schedConfig.IsNull()

	if enabled && !configSet {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("sched_config"),
			"Missing sched_config",
			"`sched_config` is required when `sched_enabled = true` (Seqera Intelligent Compute mode). "+
				"Provide a `sched_config` block with `provisioning_model` (and optionally `machine_types`), "+
				"or set `sched_enabled = false` for Classic mode.",
		)
		return
	}

	if !enabled && configSet {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("sched_config"),
			"Unexpected sched_config",
			"`sched_config` must be omitted when `sched_enabled = false` (Classic mode). "+
				"Remove the `sched_config` block, or set `sched_enabled = true` to enable Seqera Intelligent Compute.",
		)
	}
}

func SchedConfigConsistencyValidator() validator.Object {
	return ObjectSchedConfigConsistencyValidatorValidator{}
}
