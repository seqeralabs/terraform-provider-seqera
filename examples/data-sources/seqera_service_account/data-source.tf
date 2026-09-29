# Bind an existing service account, created outside Terraform, to a new agent.
data "seqera_service_account" "ci" {
  org_id = var.org_id
  name   = "ci-agent"
}

resource "seqera_agent" "nightly" {
  workspace_id       = var.workspace_id
  name               = "nightly-report"
  agent_instructions = "Summarise last night's runs."
  service_account_id = data.seqera_service_account.ci.id
}
