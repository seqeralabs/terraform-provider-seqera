# Runs an agent each time a run of the watched pipeline fails, instead of
# launching a pipeline. An agent action sets response_type = "agent" and the
# agent to run, and must not set launch or pipeline. The source must be
# pipeline_status, cron or bucket.

# The agent runs as a service account, which must be a workspace participant
# with a role that allows launching agents.
resource "seqera_service_account" "triage" {
  org_id = seqera_workspace.main.org_id
  name   = "run-triage"
}

resource "seqera_workspace_participant" "triage" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id
  member_id    = seqera_service_account.triage.member_id
  role         = "launch"
}

resource "seqera_agent" "triage" {
  workspace_id       = seqera_workspace.main.id
  name               = "failed-run-triage"
  agent_instructions = "Read the logs of the failed run, summarise the root cause and suggest a fix."
  service_account_id = seqera_service_account.triage.id

  depends_on = [seqera_workspace_participant.triage]
}

resource "seqera_action" "triage_on_failure" {
  workspace_id  = seqera_workspace.main.id
  name          = "triage-failed-runs"
  source        = "pipeline_status"
  response_type = "agent"

  pipeline_status = {
    pipeline_id = seqera_pipeline.upstream.pipeline_id
    run_status  = "FAILED"
  }

  agent = {
    agent_config_id = seqera_agent.triage.id
  }
}
