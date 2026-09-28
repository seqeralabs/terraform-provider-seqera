resource "seqera_agent" "my_agent" {
  agent_instructions       = "...my_agent_instructions..."
  description              = "...my_description..."
  github_app_credential_id = "...my_github_app_credential_id..."
  name                     = "...my_name..."
  service_account_id       = 7
  workspace_id             = 10
}