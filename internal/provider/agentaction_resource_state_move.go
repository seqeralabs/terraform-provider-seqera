// This file implements ResourceWithMoveState for AgentActionResource so that
// an agent action managed by seqera_action (response_type = "agent") can move
// to seqera_agent_action without being recreated.
//
// This is a sidecar file that adds the MoveState method to the generated
// AgentActionResource type. Speakeasy does not manage this file.
//
// Usage in Terraform configuration (Terraform 1.8+):
//
//	moved {
//	  from = seqera_action.example
//	  to   = seqera_agent_action.example
//	}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithMoveState = &AgentActionResource{}

func (r *AgentActionResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				moveActionState(ctx, req, resp, r, "seqera_agent_action", "agent", hoistAgentID)
			},
		},
	}
}

// hoistAgentID moves seqera_action's `agent.agent_config_id` to the
// root-level `agent_id` of seqera_agent_action.
func hoistAgentID(rawState map[string]interface{}) {
	if agent, ok := rawState["agent"].(map[string]interface{}); ok {
		rawState["agent_id"] = agent["agent_config_id"]
	}
}
