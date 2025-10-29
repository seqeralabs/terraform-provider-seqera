# Gitea Credentials Example
# These are example non-sensitive values for testing

# Gitea credential
resource "seqera_gitea_credential" "example" {
  name     = "Example Gitea Credentials"
  username = "example-user"
  password = "example-password-or-token-123456"
  base_url = "https://gitea.example.com"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "gitea_credential_id" {
  value       = seqera_gitea_credential.example.credentials_id
  description = "The ID of the Gitea credential"
}

output "gitea_credential_provider_type" {
  value       = seqera_gitea_credential.example.provider_type
  description = "The provider type (should be 'gitea')"
}
