package provider

import (
	"encoding/json"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

// value_version is a Terraform-only rotation trigger and must never reach the API.
func TestPipelineSecretRequestsOmitValueVersion(t *testing.T) {
	value := "s3cret"
	version := int64(2)

	for name, req := range map[string]json.Marshaler{
		"create": shared.CreatePipelineSecretRequest{Name: "NAME", Value: value, ValueVersion: &version},
		"update": shared.UpdatePipelineSecretRequest{Value: &value, ValueVersion: &version},
	} {
		out, err := req.MarshalJSON()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var m map[string]any
		if err := json.Unmarshal(out, &m); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if _, ok := m["valueVersion"]; ok {
			t.Errorf("%s: valueVersion leaked into request body: %s", name, out)
		}
		if m["value"] != value {
			t.Errorf("%s: value missing from request body: %s", name, out)
		}
	}
}
