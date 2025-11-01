# Tower Agent Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variable for sensitive credentials
variable "agent_connection_id" {
  description = "Tower Agent connection ID"
  type        = string
  sensitive   = true
}

# Example: Basic Tower Agent credentials
resource "seqera_tower_agent_credential" "example" {
  name         = "agent-main"
  workspace_id = seqera_workspace.main.id

  connection_id = var.agent_connection_id
}
