# Bitbucket Credentials Example
# These are example non-sensitive values for testing

# Bitbucket credential for Bitbucket Cloud
resource "seqera_bitbucket_credential" "example_cloud" {
  name     = "Example Bitbucket Cloud Credentials"
  username = "example-user@example.com"
  token    = "ATBBExampleAppPassword123456789ABCDEFGH"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Bitbucket credential for self-hosted Bitbucket Server
resource "seqera_bitbucket_credential" "example_server" {
  name     = "Example Bitbucket Server Credentials"
  username = "example-user"
  token    = "ATBBExampleAppPassword123456789ABCDEFGH"
  base_url = "https://bitbucket.example.com"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "bitbucket_credential_id" {
  value       = seqera_bitbucket_credential.example_cloud.credentials_id
  description = "The ID of the Bitbucket credential"
}

output "bitbucket_credential_provider_type" {
  value       = seqera_bitbucket_credential.example_cloud.provider_type
  description = "The provider type (should be 'bitbucket')"
}
