# Tower Agent Credentials Example
# These are example non-sensitive values for testing

# Tower Agent credential (private)
resource "seqera_tower_agent_credential" "example_basic" {
  name          = "example-tower-agent-credentials"
  connection_id = "12345678-1234-1234-1234-123456789012"
  shared        = false

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Tower Agent credential (shared with workspace)
resource "seqera_tower_agent_credential" "example_shared" {
  name          = "example-tower-agent-credentials-shared"
  connection_id = "87654321-4321-4321-4321-210987654321"
  shared        = true

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "tower_agent_credential_id" {
  value       = seqera_tower_agent_credential.example_basic.credentials_id
  description = "The ID of the Tower Agent credential"
}

output "tower_agent_credential_provider_type" {
  value       = seqera_tower_agent_credential.example_basic.provider_type
  description = "The provider type (should be 'tower-agent')"
}
