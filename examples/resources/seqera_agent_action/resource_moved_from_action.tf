# Moves an agent action managed by seqera_action (response_type = "agent") to
# seqera_agent_action without recreating it. Requires Terraform 1.8 or later.
# Replace the seqera_action block with the seqera_agent_action one:
# agent.agent_config_id becomes agent_id, and response_type goes away.
moved {
  from = seqera_action.triage_on_failure
  to   = seqera_agent_action.triage_on_failure
}

resource "seqera_agent_action" "triage_on_failure" {
  workspace_id = seqera_workspace.main.id
  name         = "triage-failed-runs"
  source       = "pipeline_status"
  agent_id     = seqera_agent.triage.id

  pipeline_status = {
    pipeline_id = seqera_pipeline.upstream.pipeline_id
    run_status  = "FAILED"
  }
}
