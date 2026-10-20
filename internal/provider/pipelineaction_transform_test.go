package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

// The jq transforms on the pipeline action schemas live in the overlay
// (overlays/actions-pipeline.yaml); these tests pin their behavior through
// the generated SDK models.

func TestPipelineActionReadTransformHoistsTrigger(t *testing.T) {
	body := `{"action": {
		"id": "act456", "name": "on-upload", "source": "bucket", "responseType": "pipeline",
		"config": {"discriminator": "bucket", "dataLinkId": "dl-1", "markerFile": "ready.txt",
		           "events": ["ObjectCreated"], "bucketName": "inputs", "topicArn": "arn:aws:sns:x"},
		"launch": {"pipeline": "https://github.com/nf-core/rnaseq", "workDir": "s3://my-bucket/work"},
		"pipeline": {"targetPipelineId": 9}
	}}`
	var resp shared.DescribePipelineActionResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatal(err)
	}
	a := resp.Action
	if a.Bucket == nil || a.Bucket.DataLinkID != "dl-1" || a.Bucket.MarkerFile == nil || *a.Bucket.MarkerFile != "ready.txt" {
		t.Errorf("bucket not hoisted from config: %+v", a.Bucket)
	}
	if a.Cron != nil || a.PipelineStatus != nil {
		t.Errorf("only the trigger matching config.discriminator should be set: cron=%+v pipelineStatus=%+v", a.Cron, a.PipelineStatus)
	}
	if a.Launch == nil || a.Launch.Pipeline == nil || *a.Launch.Pipeline != "https://github.com/nf-core/rnaseq" {
		t.Errorf("launch not decoded: %+v", a.Launch)
	}
}

func TestPipelineActionReadTransformTowerHasNoTrigger(t *testing.T) {
	body := `{"action": {
		"id": "act456", "name": "api", "source": "tower",
		"config": {"discriminator": "tower"}, "hookUrl": "https://example.com/hook"
	}}`
	var resp shared.DescribePipelineActionResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("an action without responseType (older Platform) must decode as a pipeline action: %v", err)
	}
	a := resp.Action
	if a.Bucket != nil || a.Cron != nil || a.PipelineStatus != nil {
		t.Errorf("a tower action has no trigger block: bucket=%+v cron=%+v pipelineStatus=%+v", a.Bucket, a.Cron, a.PipelineStatus)
	}
	if a.HookURL == nil || *a.HookURL != "https://example.com/hook" {
		t.Errorf("hookUrl = %v", a.HookURL)
	}
}

func TestPipelineActionReadTransformRefusesAgentActions(t *testing.T) {
	body := `{"action": {"id": "act123", "name": "triage", "source": "cron", "responseType": "agent"}}`
	var resp shared.DescribePipelineActionResponse
	err := json.Unmarshal([]byte(body), &resp)
	if err == nil {
		t.Fatal("expected decoding an agent action as a pipeline action to fail")
	}
	if !strings.Contains(err.Error(), "seqera_agent_action") {
		t.Errorf("error should point to seqera_agent_action, got: %v", err)
	}
}

func TestPipelineActionCreateTransformSendsPipelineOutcome(t *testing.T) {
	out, err := json.Marshal(shared.CreatePipelineActionRequest{Name: "api"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["responseType"] != "pipeline" {
		t.Errorf("responseType = %v, want pipeline", got["responseType"])
	}
	if _, present := got["agent"]; present {
		t.Error("a pipeline action must not send agent")
	}
}
