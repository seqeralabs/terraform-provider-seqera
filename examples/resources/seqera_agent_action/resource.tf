# Runs an agent each time a run of the watched pipeline fails. The agent runs
# as a service account, which must be a workspace participant with a role that
# allows launching agents.
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
