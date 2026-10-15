// This file implements ResourceWithMoveState for PipelineActionResource so
// that a pipeline action managed by seqera_action (response_type = "pipeline"
// or unset) can move to seqera_pipeline_action without being recreated.
//
// This is a sidecar file that adds the MoveState method to the generated
// PipelineActionResource type. Speakeasy does not manage this file.
//
// Usage in Terraform configuration (Terraform 1.8+):
//
//	moved {
//	  from = seqera_action.example
//	  to   = seqera_pipeline_action.example
//	}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithMoveState = &PipelineActionResource{}

func (r *PipelineActionResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			// launch, pipeline and the trigger blocks keep their names and
			// shape; everything seqera_action-only is dropped by the decode.
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				moveActionState(ctx, req, resp, r, "seqera_pipeline_action", "pipeline", nil)
			},
		},
	}
}
