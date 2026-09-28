package stringplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.String = StringServiceAccountNameFollowsIDPlanModifier{}

// StringServiceAccountNameFollowsIDPlanModifier keeps seqera_agent's
// service_account_name from state while service_account_id is unchanged.
// The name is derived from the bound service account, so it can only change
// when the binding does; otherwise every update would show it as
// "(known after apply)".
type StringServiceAccountNameFollowsIDPlanModifier struct{}

// Description describes the plan modification in plain text formatting.
func (v StringServiceAccountNameFollowsIDPlanModifier) Description(_ context.Context) string {
	return "Keeps the prior value unless service_account_id changes."
}

// MarkdownDescription describes the plan modification in Markdown formatting.
func (v StringServiceAccountNameFollowsIDPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Keeps the prior value unless `service_account_id` changes."
}

// PlanModifyString performs the plan modification.
func (v StringServiceAccountNameFollowsIDPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Nothing to do on create or destroy, or when the plan already has a value.
	if !req.PlanValue.IsUnknown() || req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var stateID, planID types.Int64
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("service_account_id"), &stateID)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("service_account_id"), &planID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// An unknown binding (e.g. a service account created in the same apply)
	// or a different one means the name may change.
	if planID.IsUnknown() || !planID.Equal(stateID) {
		return
	}
	resp.PlanValue = req.StateValue
}

func ServiceAccountNameFollowsID() planmodifier.String {
	return StringServiceAccountNameFollowsIDPlanModifier{}
}
