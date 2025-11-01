variable "agent_connection_id" {
  type      = string
  sensitive = true
}

resource "seqera_tower_agent_credential" "example" {
  name         = "agent-main"
  workspace_id = seqera_workspace.main.id

  connection_id = var.agent_connection_id
}
