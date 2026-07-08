package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/seqeralabs/terraform-provider-seqera/internal/stateupgraders"
)

// computeEnvUpgraders are the two registered seqera_compute_env upgraders. The
// framework does not chain upgraders, so both must migrate their version directly
// to the current schema. These tests run against the real schema (registered by
// this package's init() in stateupgrader_schemas.go), which is the only place the
// removed-attribute handling can be exercised faithfully.
var computeEnvUpgraders = map[string]func(context.Context, resource.UpgradeStateRequest, *resource.UpgradeStateResponse){
	"v0": stateupgraders.ComputeenvStateUpgraderV0,
	"v1": stateupgraders.ComputeenvStateUpgraderV1,
}

// runUpgraderAgainstSchema runs an upgrader over priorState and returns the
// upgraded state decoded as a tfsdk.State against the current schema of the
// resource built by newResource. The decode uses the framework's own strict
// DynamicValue.Unmarshal (no IgnoreUndefinedAttributes) — the exact check that
// rejected old state before this fix — so any attribute the upgrader failed to
// drop fails the test here.
func runUpgraderAgainstSchema(
	t *testing.T,
	newResource func() resource.Resource,
	upgrade func(context.Context, resource.UpgradeStateRequest, *resource.UpgradeStateResponse),
	priorState map[string]interface{},
) tfsdk.State {
	t.Helper()
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	newResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema errors: %v", schemaResp.Diagnostics)
	}
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)

	priorJSON, err := json.Marshal(priorState)
	if err != nil {
		t.Fatalf("marshal prior state: %v", err)
	}

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: priorJSON}}
	resp := &resource.UpgradeStateResponse{}
	upgrade(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrader diagnostics: %v", resp.Diagnostics)
	}
	if resp.DynamicValue == nil {
		t.Fatalf("upgrader returned no DynamicValue")
	}

	value, err := resp.DynamicValue.Unmarshal(schemaType)
	if err != nil {
		t.Fatalf("upgraded state failed to decode against v2 schema: %v", err)
	}

	return tfsdk.State{Schema: schemaResp.Schema, Raw: value}
}

// TestComputeenvUpgrade_DropsRemovedAttributesAndPreservesData reproduces issue
// #228: prior state carries attributes removed from the v2 schema — the
// top-level `deleted` and (this is what the synthetic fixtures originally missed)
// per-platform fields like `config.aws_cloud.enable_fusion`. The upgraded state
// must decode cleanly against v2 (removed attributes dropped) while preserving
// the attributes that still exist.
func TestComputeenvUpgrade_DropsRemovedAttributesAndPreservesData(t *testing.T) {
	ctx := context.Background()

	for name, upgrade := range computeEnvUpgraders {
		t.Run(name, func(t *testing.T) {
			state := runUpgraderAgainstSchema(t, NewComputeEnvResource, upgrade, map[string]interface{}{
				"id":             "ce-123",
				"compute_env_id": "ce-123",
				"compute_env": map[string]interface{}{
					"compute_env_id": "ce-123",
					"name":           "compute_spot_fusion",
					"deleted":        false, // removed in v2
					"credentials_id": "cred-1",
					"platform":       "aws-cloud",
					"config": map[string]interface{}{
						"aws_cloud": map[string]interface{}{
							"region":        "eu-west-2", // still present in v2
							"work_dir":      "s3://bucket/work",
							"enable_fusion": true, // removed from aws_cloud in v2
							"enable_wave":   true, // removed from aws_cloud in v2
							"ebs_boot_size": 50,   // removed from aws_cloud in v2
						},
					},
				},
			})

			// Decoding succeeded (removals dropped). Confirm a surviving attribute
			// was carried across unchanged.
			var region types.String
			diags := state.GetAttribute(ctx, path.Root("compute_env").AtName("config").AtName("aws_cloud").AtName("region"), &region)
			if diags.HasError() {
				t.Fatalf("reading region: %v", diags)
			}
			if region.ValueString() != "eu-west-2" {
				t.Errorf("expected region preserved as eu-west-2, got %q", region.ValueString())
			}
		})
	}
}

// TestComputeenvUpgrade_DerivesAzureDeleteJobsOnCompletion covers the value
// transform: the legacy `delete_jobs_on_completion` string becomes the boolean
// `delete_jobs_on_completion_enabled`.
func TestComputeenvUpgrade_DerivesAzureDeleteJobsOnCompletion(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name  string
		state map[string]interface{}
		want  *bool
	}{
		{"on_success -> true", azureBatchState("on_success", nil), boolPtr(true)},
		{"always -> true", azureBatchState("always", nil), boolPtr(true)},
		{"never -> false", azureBatchState("never", nil), boolPtr(false)},
		{"explicit false preserved", azureBatchState("always", boolPtr(false)), boolPtr(false)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Exercise the v1 upgrader; the v0 upgrader shares the same transform.
			state := runUpgraderAgainstSchema(t, NewComputeEnvResource, stateupgraders.ComputeenvStateUpgraderV1, tc.state)

			var enabled types.Bool
			diags := state.GetAttribute(ctx, path.Root("compute_env").AtName("config").AtName("azure_batch").AtName("delete_jobs_on_completion_enabled"), &enabled)
			if diags.HasError() {
				t.Fatalf("reading delete_jobs_on_completion_enabled: %v", diags)
			}
			if enabled.IsNull() {
				t.Fatalf("expected delete_jobs_on_completion_enabled to be set, got null")
			}
			if enabled.ValueBool() != *tc.want {
				t.Errorf("expected delete_jobs_on_completion_enabled=%v, got %v", *tc.want, enabled.ValueBool())
			}
		})
	}
}

// TestComputeenvUpgrade_V0RenamesNvmeFlag covers the v0->v1 rename of the
// misspelled `nvnme_storage_enabled` flag, carrying its value to the correctly
// spelled attribute.
func TestComputeenvUpgrade_V0RenamesNvmeFlag(t *testing.T) {
	ctx := context.Background()

	state := runUpgraderAgainstSchema(t, NewComputeEnvResource, stateupgraders.ComputeenvStateUpgraderV0, map[string]interface{}{
		"id":             "ce-123",
		"compute_env_id": "ce-123",
		"compute_env": map[string]interface{}{
			"compute_env_id": "ce-123",
			"name":           "compute_spot_fusion",
			"credentials_id": "cred-1",
			"platform":       "aws-batch",
			"config": map[string]interface{}{
				"aws_batch": map[string]interface{}{
					"region":                "eu-west-2",
					"work_dir":              "s3://bucket/work",
					"nvnme_storage_enabled": true, // misspelled v0 name
				},
			},
		},
	})

	var nvme types.Bool
	diags := state.GetAttribute(ctx, path.Root("compute_env").AtName("config").AtName("aws_batch").AtName("nvme_storage_enabled"), &nvme)
	if diags.HasError() {
		t.Fatalf("reading nvme_storage_enabled: %v", diags)
	}
	if nvme.IsNull() || !nvme.ValueBool() {
		t.Errorf("expected nvme_storage_enabled=true after rename, got %v", nvme)
	}
}

func azureBatchState(deleteJobsOnCompletion string, enabled *bool) map[string]interface{} {
	azureBatch := map[string]interface{}{
		"delete_jobs_on_completion": deleteJobsOnCompletion,
	}
	if enabled != nil {
		azureBatch["delete_jobs_on_completion_enabled"] = *enabled
	}
	return map[string]interface{}{
		"id":             "ce-123",
		"compute_env_id": "ce-123",
		"compute_env": map[string]interface{}{
			"compute_env_id": "ce-123",
			"name":           "azure_ce",
			"credentials_id": "cred-1",
			"platform":       "azure-batch",
			"deleted":        false,
			"config": map[string]interface{}{
				"azure_batch": azureBatch,
			},
		},
	}
}

func boolPtr(b bool) *bool { return &b }

// TestAwsComputeEnvUpgrade_RenamesNvmeAndDropsFusion covers the confirmed-broken
// legacy seqera_aws_compute_env migration. Real v0.25.x state carries
// config.{nvnme_storage_enabled (misspelled), fusion_enabled, fusion2_enabled,
// wave_enabled}. The current schema renamed the nvme flag and removed the three
// fusion flags (replaced by enable_fusion/enable_wave). The upgrader must carry
// the nvme value to the new name, drop the removed flags, and preserve the rest.
func TestAwsComputeEnvUpgrade_RenamesNvmeAndDropsFusion(t *testing.T) {
	ctx := context.Background()

	state := runUpgraderAgainstSchema(t, NewAWSComputeEnvResource, stateupgraders.AwscomputeenvStateUpgraderV0, map[string]interface{}{
		"id":             "ce-1",
		"compute_env_id": "ce-1",
		"name":           "legacy-aws-ce",
		"platform":       "aws-batch",
		"config": map[string]interface{}{
			"region":                "eu-west-2", // preserved
			"work_dir":              "s3://bucket/work",
			"nvnme_storage_enabled": true, // misspelled v0 name -> renamed
			"fusion_enabled":        true, // removed in current -> dropped
			"fusion2_enabled":       true, // removed in current -> dropped
			"wave_enabled":          true, // removed in current -> dropped
		},
	})

	// Decoding succeeded, so the removed fusion flags were dropped. Confirm the
	// nvme rename carried its value and a surviving attribute is intact.
	var nvme types.Bool
	if diags := state.GetAttribute(ctx, path.Root("config").AtName("nvme_storage_enabled"), &nvme); diags.HasError() {
		t.Fatalf("reading nvme_storage_enabled: %v", diags)
	}
	if nvme.IsNull() || !nvme.ValueBool() {
		t.Errorf("expected nvme_storage_enabled=true after rename, got %v", nvme)
	}

	var region types.String
	if diags := state.GetAttribute(ctx, path.Root("config").AtName("region"), &region); diags.HasError() {
		t.Fatalf("reading region: %v", diags)
	}
	if region.ValueString() != "eu-west-2" {
		t.Errorf("expected region preserved as eu-west-2, got %q", region.ValueString())
	}
}

// TestAwscredentialUpgrade_DropsRemovedAttributes covers the passthrough-style
// upgrader on a non-compute-env resource. Real v0.25.x seqera_aws_credential
// state carries attributes since removed from the schema (`deleted`, `keys`,
// `date_created`, `base_url`, `category`, …). The upgrader must drop them so the
// state decodes against the current schema, while preserving what remains.
func TestAwscredentialUpgrade_DropsRemovedAttributes(t *testing.T) {
	ctx := context.Background()

	state := runUpgraderAgainstSchema(t, NewAWSCredentialResource, stateupgraders.AwscredentialStateUpgraderV0, map[string]interface{}{
		"id":             "cred-1",
		"credentials_id": "cred-1",
		"name":           "my-cred", // still present in the current schema
		// Attributes removed from the schema since v0.25.x:
		"deleted":      false,
		"date_created": "2024-01-01T00:00:00Z",
		"last_used":    nil,
		"last_updated": nil,
		"base_url":     nil,
		"category":     nil,
		"checked":      false,
		"description":  nil,
		"keys": map[string]interface{}{
			"access_key": "AKIAEXAMPLE",
			"secret_key": "secret",
		},
	})

	var name types.String
	if diags := state.GetAttribute(ctx, path.Root("name"), &name); diags.HasError() {
		t.Fatalf("reading name: %v", diags)
	}
	if name.ValueString() != "my-cred" {
		t.Errorf("expected name preserved as my-cred, got %q", name.ValueString())
	}
}
