// This file implements ResourceWithMoveState for AzureCloudCEResource
// to allow migration from compatible compute environment resources to
// seqera_azure_cloud_ce without destroying and recreating the resource.
//
// This is a sidecar file that adds the MoveState method to the generated
// AzureCloudCEResource type. Speakeasy does not manage this file.
//
// Supported source resource types:
//   - seqera_compute_env
//
// Usage in Terraform configuration:
//
//	moved {
//	  from = seqera_compute_env.example
//	  to   = seqera_azure_cloud_ce.example
//	}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the resource implements ResourceWithMoveState
var _ resource.ResourceWithMoveState = &AzureCloudCEResource{}

// MoveState returns the state movers for migrating from other resource types
func (r *AzureCloudCEResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			// Allow migration from seqera_compute_env to seqera_azure_cloud_ce.
			// The schemas are compatible, so we can directly copy the state.
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				tflog.Debug(ctx, "Processing state move request", map[string]interface{}{
					"source_type":     req.SourceTypeName,
					"source_provider": req.SourceProviderAddress,
					"target_type":     "seqera_azure_cloud_ce",
				})

				supportedSources := []string{
					"seqera_compute_env",
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

				if req.SourceRawState == nil {
					tflog.Warn(ctx, "Source raw state is nil")
					return
				}

				var schemaResp resource.SchemaResponse
				r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

				if schemaResp.Diagnostics.HasError() {
					resp.Diagnostics.Append(schemaResp.Diagnostics...)
					return
				}

				rawStateValue, err := req.SourceRawState.Unmarshal(schemaResp.Schema.Type().TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Failed to unmarshal source state",
						"Could not unmarshal raw state into schema type: "+err.Error(),
					)
					return
				}

				targetState := tfsdk.State{
					Schema: schemaResp.Schema,
					Raw:    rawStateValue,
				}

				resp.TargetState = targetState

				tflog.Info(ctx, "Successfully moved state to seqera_azure_cloud_ce", map[string]interface{}{
					"source_type": req.SourceTypeName,
				})
			},
		},
	}
}
