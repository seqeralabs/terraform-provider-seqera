variable "agent_connection_id" {
  type      = string
  sensitive = true
}

resource "seqera_tower_agent_credential" "shared" {
  name         = "agent-shared"
  workspace_id = seqera_workspace.main.id

  connection_id = var.agent_connection_id
  shared        = true
}
