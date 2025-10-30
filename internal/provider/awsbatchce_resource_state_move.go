// This file implements ResourceWithMoveState for AWSBatchCEResource
// to allow migration from compatible compute environment resources to
// seqera_aws_batch_ce without destroying and recreating the resource.
//
// This is a sidecar file that adds the MoveState method to the generated
// AWSBatchCEResource type. Speakeasy does not manage this file.
//
// Supported source resource types:
//   - seqera_compute_env
//   - seqera_aws_compute_env
//
// Usage in Terraform configuration:
//
//	moved {
//	  from = seqera_compute_env.example
//	  to   = seqera_aws_batch_ce.example
//	}
//
//	moved {
//	  from = seqera_aws_compute_env.example
//	  to   = seqera_aws_batch_ce.example
//	}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the resource implements ResourceWithMoveState
var _ resource.ResourceWithMoveState = &AWSBatchCEResource{}

// MoveState returns the state movers for migrating from other resource types
func (r *AWSBatchCEResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			// Allow migration from seqera_compute_env to seqera_aws_batch_ce
			// The schemas are compatible, so we can directly copy the state
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				// Log the move state request
				tflog.Debug(ctx, "Processing state move request", map[string]interface{}{
					"source_type":     req.SourceTypeName,
					"source_provider": req.SourceProviderAddress,
					"target_type":     "seqera_aws_batch_ce",
				})

				// Handle moves from compatible compute environment resources
				// All AWS compute environment resources use the same underlying schema
				supportedSources := []string{
					"seqera_compute_env",
					"seqera_aws_compute_env",
				}

				isSupported := false
				for _, source := range supportedSources {
					if req.SourceTypeName == source {
						isSupported = true
						break
					}
				}

				if !isSupported {
					tflog.Debug(ctx, "Skipping state move: source type not supported", map[string]interface{}{
						"supported": supportedSources,
						"actual":    req.SourceTypeName,
					})
					return
				}

				// Verify we have source raw state
				if req.SourceRawState == nil {
					tflog.Warn(ctx, "Source raw state is nil")
					return
				}

				// Get the schema for the target resource to properly unmarshal the state
				var schemaResp resource.SchemaResponse
				r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

				if schemaResp.Diagnostics.HasError() {
					resp.Diagnostics.Append(schemaResp.Diagnostics...)
					return
				}

				// Unmarshal the raw state into a tftypes.Value using the target schema
				rawStateValue, err := req.SourceRawState.Unmarshal(schemaResp.Schema.Type().TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Failed to unmarshal source state",
						"Could not unmarshal raw state into schema type: "+err.Error(),
					)
					return
				}

				// Create the target state using the unmarshaled value
				targetState := tfsdk.State{
					Schema: schemaResp.Schema,
					Raw:    rawStateValue,
				}

				// Copy the state to the response
				resp.TargetState = targetState

				tflog.Info(ctx, "Successfully moved state to seqera_aws_batch_ce", map[string]interface{}{
					"source_type": req.SourceTypeName,
				})
			},
		},
	}
}
