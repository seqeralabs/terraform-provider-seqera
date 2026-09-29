package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = ObjectActionOutcomeValidator{}

// ObjectActionOutcomeValidator checks seqera_action's outcome blocks against
// response_type, mirroring the Platform's outcome planners: a pipeline
// response (the default) needs `launch` and takes no `agent`; an agent
// response needs `agent` and takes neither `launch` nor `pipeline`.
// Attached to `launch`, which is optional so that agent actions can omit it.
type ObjectActionOutcomeValidator struct{}

func (v ObjectActionOutcomeValidator) Description(_ context.Context) string {
	return "Validates that the action's outcome blocks match response_type."
}

func (v ObjectActionOutcomeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ObjectActionOutcomeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	var responseType types.String
	var agent, pipeline types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("response_type"), &responseType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("agent"), &agent)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("pipeline"), &pipeline)...)
	if resp.Diagnostics.HasError() || responseType.IsUnknown() {
		return
	}

	isSet := func(o types.Object) bool { return !o.IsNull() && !o.IsUnknown() }
	launchSet := isSet(req.ConfigValue)

	if responseType.ValueString() == "agent" {
		if launchSet {
			resp.Diagnostics.AddAttributeError(path.Root("launch"), "Launch not allowed for an agent action",
				"An action with `response_type = \"agent\"` runs the agent instead of a pipeline, so it must not set `launch`.")
		}
		if isSet(pipeline) {
			resp.Diagnostics.AddAttributeError(path.Root("pipeline"), "Pipeline not allowed for an agent action",
				"An action with `response_type = \"agent\"` runs the agent instead of a pipeline, so it must not set `pipeline`.")
		}
		if agent.IsNull() {
			resp.Diagnostics.AddAttributeError(path.Root("agent"), "Missing agent",
				"An action with `response_type = \"agent\"` must set `agent.agent_config_id` to the agent it runs.")
		}
		return
	}

	// A null response_type defaults to "pipeline".
	if req.ConfigValue.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("launch"), "Missing launch",
			"An action that launches a pipeline (`response_type = \"pipeline\"`, the default) must set `launch`.")
	}
	if isSet(agent) {
		resp.Diagnostics.AddAttributeError(path.Root("agent"), "Agent not allowed for a pipeline action",
			"`agent` is only allowed when `response_type = \"agent\"`.")
	}
}

func ActionOutcomeValidator() validator.Object {
	return ObjectActionOutcomeValidator{}
}
