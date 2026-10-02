# Look up an agent managed outside Terraform.
data "seqera_agent" "triage" {
  workspace_id = var.workspace_id
  name         = "failed-run-triage"
}

output "triage_agent_id" {
  value = data.seqera_agent.triage.id
}

output "triage_agent_status" {
  value = data.seqera_agent.triage.status
}
