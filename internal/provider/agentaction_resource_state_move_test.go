package provider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// seqeraActionState returns a seqera_action raw state with every attribute of
// the current seqera_action schema present (null unless overridden), so the
// move is exercised against all the attributes seqera_agent_action lacks.
func seqeraActionState(t *testing.T, overrides map[string]interface{}) *tfprotov6.RawState {
	t.Helper()
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	NewActionResource().Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("seqera_action schema errors: %v", schemaResp.Diagnostics)
	}

	state := map[string]interface{}{}
	for name := range schemaResp.Schema.Attributes {
		state[name] = nil
	}
	for k, v := range overrides {
		if _, ok := state[k]; !ok {
			t.Fatalf("fixture sets %q, which seqera_action does not define", k)
		}
		state[k] = v
	}

	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	return &tfprotov6.RawState{JSON: raw}
}

func agentActionSourceState(t *testing.T, responseType interface{}) *tfprotov6.RawState {
	return seqeraActionState(t, map[string]interface{}{
		"action_id":     "act123",
		"id":            "act123",
		"workspace_id":  42,
		"name":          "triage-failed-runs",
		"source":        "pipeline_status",
		"response_type": responseType,
		"status":        "ACTIVE",
		"agent":         map[string]interface{}{"agent_config_id": "agent-1"},
		"pipeline_status": map[string]interface{}{
			"pipeline_id": 7,
			"run_status":  "FAILED",
		},
		"config": map[string]interface{}{
			"pipeline_status": map[string]interface{}{"discriminator": "pipeline_status", "pipeline_id": 7, "run_status": "FAILED"},
		},
		"label_ids": []interface{}{1, 2},
	})
}

func runAgentActionMove(t *testing.T, sourceType string, raw *tfprotov6.RawState) *resource.MoveStateResponse {
	t.Helper()
	ctx := context.Background()
	r := &AgentActionResource{}
	movers := r.MoveState(ctx)
	if len(movers) != 1 {
		t.Fatalf("expected 1 state mover, got %d", len(movers))
	}
	resp := &resource.MoveStateResponse{}
	movers[0].StateMover(ctx, resource.MoveStateRequest{
		SourceTypeName:        sourceType,
		SourceProviderAddress: "registry.terraform.io/seqeralabs/seqera",
		SourceSchemaVersion:   1,
		SourceRawState:        raw,
	}, resp)
	return resp
}

func TestAgentActionMoveStateFromAgentAction(t *testing.T) {
	ctx := context.Background()
	resp := runAgentActionMove(t, "seqera_action", agentActionSourceState(t, "agent"))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var got AgentActionResourceModel
	if diags := resp.TargetState.Get(ctx, &got); diags.HasError() {
		t.Fatalf("moved state does not decode into the model: %v", diags)
	}

	checks := map[string][2]interface{}{
		"id":           {got.ID, types.StringValue("act123")},
		"workspace_id": {got.WorkspaceID, types.Int64Value(42)},
		"name":         {got.Name, types.StringValue("triage-failed-runs")},
		"source":       {got.Source, types.StringValue("pipeline_status")},
		"agent_id":     {got.AgentID, types.StringValue("agent-1")},
		"status":       {got.Status, types.StringValue("ACTIVE")},
	}
	for name, c := range checks {
		if c[0] != c[1] {
			t.Errorf("%s = %v, want %v", name, c[0], c[1])
		}
	}
	if got.PipelineStatus == nil ||
		got.PipelineStatus.PipelineID != types.Int64Value(7) ||
		got.PipelineStatus.RunStatus != types.StringValue("FAILED") {
		t.Errorf("pipeline_status not carried over: %+v", got.PipelineStatus)
	}
	if got.Bucket != nil || got.Cron != nil {
		t.Errorf("unset trigger blocks should stay null, got bucket=%+v cron=%+v", got.Bucket, got.Cron)
	}
	if len(got.LabelIds) != 2 {
		t.Errorf("label_ids not carried over: %v", got.LabelIds)
	}
}

func TestAgentActionMoveStateFallsBackToActionID(t *testing.T) {
	ctx := context.Background()
	raw := seqeraActionState(t, map[string]interface{}{
		"action_id":     "act123",
		"name":          "n",
		"source":        "cron",
		"response_type": "agent",
		"agent":         map[string]interface{}{"agent_config_id": "agent-1"},
	})
	resp := runAgentActionMove(t, "seqera_action", raw)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	var got AgentActionResourceModel
	if diags := resp.TargetState.Get(ctx, &got); diags.HasError() {
		t.Fatal(diags)
	}
	if got.ID != types.StringValue("act123") {
		t.Errorf("id = %v, want act123 from action_id", got.ID)
	}
}

func TestAgentActionMoveStateRefusesPipelineActions(t *testing.T) {
	for name, responseType := range map[string]interface{}{
		"explicit pipeline":         "pipeline",
		"missing defaults pipeline": nil,
	} {
		t.Run(name, func(t *testing.T) {
			resp := runAgentActionMove(t, "seqera_action", agentActionSourceState(t, responseType))
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected an error moving a pipeline action to seqera_agent_action")
			}
			if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, "seqera_pipeline_action") {
				t.Errorf("error should point to seqera_pipeline_action, got: %s", detail)
			}
			if !resp.TargetState.Raw.IsNull() {
				t.Error("no target state should be set on error")
			}
		})
	}
}

func TestAgentActionMoveStateIgnoresOtherSources(t *testing.T) {
	resp := runAgentActionMove(t, "seqera_compute_env", agentActionSourceState(t, "agent"))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unsupported source types must be skipped silently, got: %v", resp.Diagnostics)
	}
	if !resp.TargetState.Raw.IsNull() {
		t.Error("no target state should be set for an unsupported source type")
	}
}
