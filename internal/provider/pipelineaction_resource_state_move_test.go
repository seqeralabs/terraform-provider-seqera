package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func pipelineActionSourceState(t *testing.T, responseType interface{}) *tfprotov6.RawState {
	return seqeraActionState(t, map[string]interface{}{
		"action_id":     "act456",
		"id":            "act456",
		"workspace_id":  42,
		"name":          "nightly-run",
		"source":        "cron",
		"response_type": responseType,
		"status":        "ACTIVE",
		"hook_url":      nil,
		"cron":          map[string]interface{}{"expression": "0 2 * * *", "timezone": "UTC", "preset": nil},
		"launch": map[string]interface{}{
			"pipeline":       "https://github.com/nextflow-io/hello",
			"compute_env_id": "ce-1",
			"work_dir":       "s3://my-bucket/work",
			"revision":       "master",
		},
		"pipeline": map[string]interface{}{"target_pipeline_id": 9, "target_pipeline_version_id": "v1"},
	})
}

func runPipelineActionMove(t *testing.T, sourceType string, raw *tfprotov6.RawState) *resource.MoveStateResponse {
	t.Helper()
	ctx := context.Background()
	r := &PipelineActionResource{}
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

func TestPipelineActionMoveStateFromPipelineAction(t *testing.T) {
	ctx := context.Background()
	for name, responseType := range map[string]interface{}{
		"explicit pipeline":         "pipeline",
		"missing defaults pipeline": nil,
	} {
		t.Run(name, func(t *testing.T) {
			resp := runPipelineActionMove(t, "seqera_action", pipelineActionSourceState(t, responseType))
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			var got PipelineActionResourceModel
			if diags := resp.TargetState.Get(ctx, &got); diags.HasError() {
				t.Fatalf("moved state does not decode into the model: %v", diags)
			}
			if got.ID != types.StringValue("act456") || got.WorkspaceID != types.Int64Value(42) ||
				got.Name != types.StringValue("nightly-run") || got.Source != types.StringValue("cron") {
				t.Errorf("identity attributes not carried over: id=%v workspace=%v name=%v source=%v",
					got.ID, got.WorkspaceID, got.Name, got.Source)
			}
			if got.Cron == nil || got.Cron.Expression != types.StringValue("0 2 * * *") {
				t.Errorf("cron not carried over: %+v", got.Cron)
			}
			if got.Launch == nil ||
				got.Launch.Pipeline != types.StringValue("https://github.com/nextflow-io/hello") ||
				got.Launch.ComputeEnvID != types.StringValue("ce-1") ||
				got.Launch.WorkDir != types.StringValue("s3://my-bucket/work") {
				t.Errorf("launch not carried over: %+v", got.Launch)
			}
			if got.Pipeline == nil || got.Pipeline.TargetPipelineID != types.Int64Value(9) {
				t.Errorf("pipeline not carried over: %+v", got.Pipeline)
			}
		})
	}
}

func TestPipelineActionMoveStateRefusesAgentActions(t *testing.T) {
	resp := runPipelineActionMove(t, "seqera_action", pipelineActionSourceState(t, "agent"))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error moving an agent action to seqera_pipeline_action")
	}
	if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, "seqera_agent_action") {
		t.Errorf("error should point to seqera_agent_action, got: %s", detail)
	}
	if !resp.TargetState.Raw.IsNull() {
		t.Error("no target state should be set on error")
	}
}

func TestPipelineActionMoveStateIgnoresOtherSources(t *testing.T) {
	resp := runPipelineActionMove(t, "seqera_agent_action", pipelineActionSourceState(t, "pipeline"))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unsupported source types must be skipped silently, got: %v", resp.Diagnostics)
	}
	if !resp.TargetState.Raw.IsNull() {
		t.Error("no target state should be set for an unsupported source type")
	}
}

// TestTypedActionSchemasMatchSeqeraAction guards the state movers: they copy
// these attributes from seqera_action as-is and drop anything the target
// schema does not define. If a regeneration changes the shape of one of them
// in only one resource, the move would silently lose values (for example a
// launch setting), so the types must stay identical.
func TestTypedActionSchemasMatchSeqeraAction(t *testing.T) {
	ctx := context.Background()
	attrTypes := func(r resource.Resource) map[string]string {
		var resp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("schema errors: %v", resp.Diagnostics)
		}
		out := map[string]string{}
		for name, a := range resp.Schema.Attributes {
			out[name] = a.GetType().TerraformType(ctx).String()
		}
		return out
	}

	action := attrTypes(NewActionResource())
	targets := map[string]struct {
		res   resource.Resource
		attrs []string
	}{
		"seqera_pipeline_action": {NewPipelineActionResource(), []string{
			"id", "workspace_id", "name", "source", "bucket", "cron", "pipeline_status", "launch", "pipeline",
			"status", "hook_id", "hook_url", "error", "next_execution",
		}},
		"seqera_agent_action": {NewAgentActionResource(), []string{
			"id", "workspace_id", "name", "source", "bucket", "cron", "pipeline_status", "label_ids",
			"status", "error", "next_execution",
		}},
	}
	for typeName, target := range targets {
		got := attrTypes(target.res)
		for _, name := range target.attrs {
			if got[name] != action[name] {
				t.Errorf("%s.%s type differs from seqera_action.%s:\n  %s\n  %s",
					typeName, name, name, got[name], action[name])
			}
		}
	}
}
