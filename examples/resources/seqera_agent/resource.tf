# Recommended setup: the agent runs as a dedicated service account.
resource "seqera_service_account" "triage" {
  org_id = var.org_id
  name   = "run-triage"
}

# The service account must be a workspace participant with a role that allows
# launching agents (`launch` or higher) before the agent is saved.
resource "seqera_workspace_participant" "triage" {
  org_id       = var.org_id
  workspace_id = var.workspace_id
  member_id    = seqera_service_account.triage.member_id
  role         = "launch"
}

resource "seqera_agent" "triage" {
  workspace_id       = var.workspace_id
  name               = "failed-run-triage"
  description        = "Explains failed pipeline runs"
  agent_instructions = <<-EOT
    When a run fails, read its logs, summarise the root cause in two
    sentences and suggest the smallest fix.
  EOT
  service_account_id = seqera_service_account.triage.id

  # Otherwise the Platform rejects the binding with a conflict.
  depends_on = [seqera_workspace_participant.triage]
}

# Without a service account the agent runs with the identity of whoever
# triggers it: for an action, the action's owner (the user behind the
# provider's token for Terraform-managed actions); for a direct launch, the
# launching user. Prefer binding a service account where they are enabled.
resource "seqera_agent" "minimal" {
  workspace_id       = var.workspace_id
  name               = "release-notes"
  agent_instructions = "Draft release notes from the merged pull requests."
}
