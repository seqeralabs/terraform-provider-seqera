package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

// The jq transforms on the agent action schemas live in the overlay
// (overlays/actions-agent.yaml); these tests pin their behavior through the
// generated SDK models.

func TestAgentActionReadTransformHoistsTriggerAndAgent(t *testing.T) {
	body := `{"action": {
		"id": "act123", "name": "triage", "source": "pipeline_status",
		"responseType": "agent", "status": "ACTIVE",
		"agent": {"agentConfigId": "agent-1"},
		"config": {"discriminator": "pipeline_status", "pipelineId": 7, "runStatus": "FAILED"},
		"launch": null, "hookUrl": null
	}}`
	var resp shared.DescribeAgentActionResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatal(err)
	}
	a := resp.Action
	if a == nil {
		t.Fatal("action not decoded")
	}
	if a.AgentID == nil || *a.AgentID != "agent-1" {
		t.Errorf("agentId = %v, want agent-1 from agent.agentConfigId", a.AgentID)
	}
	if a.PipelineStatus == nil || a.PipelineStatus.PipelineID != 7 || string(a.PipelineStatus.RunStatus) != "FAILED" {
		t.Errorf("pipelineStatus not hoisted from config: %+v", a.PipelineStatus)
	}
	if a.Bucket != nil || a.Cron != nil {
		t.Errorf("only the trigger matching config.discriminator should be set: bucket=%+v cron=%+v", a.Bucket, a.Cron)
	}
}

func TestAgentActionReadTransformHoistsCron(t *testing.T) {
	body := `{"action": {
		"id": "act123", "name": "nightly", "source": "cron", "responseType": "agent",
		"agent": {"agentConfigId": "agent-1"},
		"config": {"discriminator": "cron", "expression": "0 2 * * *", "timezone": "UTC"}
	}}`
	var resp shared.DescribeAgentActionResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatal(err)
	}
	c := resp.Action.Cron
	if c == nil || c.Expression == nil || *c.Expression != "0 2 * * *" || c.Timezone == nil || *c.Timezone != "UTC" {
		t.Errorf("cron not hoisted from config: %+v", c)
	}
	if resp.Action.PipelineStatus != nil {
		t.Errorf("pipelineStatus should be null for a cron action: %+v", resp.Action.PipelineStatus)
	}
}

func TestAgentActionReadTransformRefusesPipelineActions(t *testing.T) {
	body := `{"action": {"id": "act123", "name": "launcher", "source": "tower", "responseType": "pipeline"}}`
	var resp shared.DescribeAgentActionResponse
	err := json.Unmarshal([]byte(body), &resp)
	if err == nil {
		t.Fatal("expected decoding a pipeline action as an agent action to fail")
	}
	if !strings.Contains(err.Error(), "seqera_pipeline_action") {
		t.Errorf("error should point to seqera_pipeline_action, got: %v", err)
	}
}

func TestAgentActionCreateTransformSendsAgentOutcome(t *testing.T) {
	agentID := "agent-1"
	req := shared.CreateAgentActionRequest{
		Name:    "triage",
		Source:  shared.CreateAgentActionRequestSourcePipelineStatus,
		AgentID: agentID,
	}
	out, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["responseType"] != "agent" {
		t.Errorf("responseType = %v, want agent", got["responseType"])
	}
	agent, _ := got["agent"].(map[string]interface{})
	if agent["agentConfigId"] != agentID {
		t.Errorf("agent.agentConfigId = %v, want %s", agent["agentConfigId"], agentID)
	}
	if _, present := got["agentId"]; present {
		t.Error("top-level agentId must not be sent to the API")
	}
}
