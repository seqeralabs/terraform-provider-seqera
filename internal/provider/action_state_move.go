// This file holds the shared state-move logic for the typed action resources
// (seqera_agent_action, seqera_pipeline_action), which accept a `moved` block
// from the generic seqera_action. Speakeasy does not manage this file.

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// actionMoveSourceType is the only resource type the typed action resources
// accept state from.
const actionMoveSourceType = "seqera_action"

// moveActionState moves a seqera_action state into a typed action resource.
//
// It refuses a state whose response_type does not match responseType (a
// missing response_type means "pipeline", the API default), applies transform
// to the raw state for the attributes whose shape differs, then decodes the
// result against the target schema with IgnoreUndefinedAttributes so every
// seqera_action-only attribute (launch, config, hook_url, ...) is dropped.
// Values the target schema computes from the API are refreshed by the next
// read; this only has to produce a decodable state.
func moveActionState(
	ctx context.Context,
	req resource.MoveStateRequest,
	resp *resource.MoveStateResponse,
	target resource.Resource,
	targetTypeName string,
	responseType string,
	transform func(rawState map[string]interface{}),
) {
	if req.SourceTypeName != actionMoveSourceType {
		tflog.Debug(ctx, "Skipping state move: source type not supported", map[string]interface{}{
			"supported": actionMoveSourceType,
			"actual":    req.SourceTypeName,
		})
		return
	}
	if req.SourceRawState == nil {
		resp.Diagnostics.AddError("Unable to Move Resource State", "The source "+actionMoveSourceType+" state is empty.")
		return
	}

	var rawState map[string]interface{}
	if err := json.Unmarshal(req.SourceRawState.JSON, &rawState); err != nil {
		resp.Diagnostics.AddError("Unable to Unmarshal Source State", err.Error())
		return
	}

	sourceResponseType, _ := rawState["response_type"].(string)
	if sourceResponseType == "" {
		sourceResponseType = "pipeline"
	}
	if sourceResponseType != responseType {
		resp.Diagnostics.AddError(
			"Action Outcome Does Not Match",
			fmt.Sprintf("The %s being moved has response_type %q and cannot be moved to %s, which only manages %q actions. Move it to seqera_%s_action instead.",
				actionMoveSourceType, sourceResponseType, targetTypeName, responseType, sourceResponseType),
		)
		return
	}

	// seqera_action keeps the identifier in both `action_id` and `id`; older
	// states may only have `action_id`. The typed resources only have `id`.
	if id, _ := rawState["id"].(string); id == "" {
		rawState["id"] = rawState["action_id"]
	}

	if transform != nil {
		transform(rawState)
	}

	var schemaResp resource.SchemaResponse
	target.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		resp.Diagnostics.Append(schemaResp.Diagnostics...)
		return
	}
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)

	movedJSON, err := json.Marshal(rawState)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Marshal Moved State", err.Error())
		return
	}
	value, err := (&tfprotov6.RawState{JSON: movedJSON}).UnmarshalWithOpts(
		schemaType,
		tfprotov6.UnmarshalOpts{
			ValueFromJSONOpts: tftypes.ValueFromJSONOpts{IgnoreUndefinedAttributes: true},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode Moved State",
			"The "+actionMoveSourceType+" state could not be decoded against the "+targetTypeName+" schema. "+
				"This is always a bug in the provider and should be reported to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	resp.TargetState = tfsdk.State{Schema: schemaResp.Schema, Raw: value}

	tflog.Info(ctx, "Moved state to "+targetTypeName, map[string]interface{}{
		"source_type": req.SourceTypeName,
	})
}
