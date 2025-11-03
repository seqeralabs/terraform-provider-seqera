variable "agent_connection_id" {
  type      = string
  sensitive = true
}

# IMPORTANT: The Tower Agent must be running and online with this connection_id
# before creating the credential. Start the agent first:
# ./tw-agent <connection_id> -u https://cloud.seqera.io/api --work-dir=/work

resource "seqera_tower_agent_credential" "example" {
  name         = "agent-main"
  workspace_id = seqera_workspace.main.id

  connection_id = var.agent_connection_id
  shared        = false
}

resource "seqera_tower_agent_credential" "shared" {
  name         = "agent-shared"
  workspace_id = seqera_workspace.main.id

  connection_id = var.agent_connection_id
  shared        = true
}
