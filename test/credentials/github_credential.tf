# GitHub Credentials Example
# These are example non-sensitive values for testing

# GitHub credential for GitHub.com
resource "seqera_github_credential" "example_github_com" {
  name         = "Example GitHub Credentials"
  access_token = "ghp_ExamplePersonalAccessToken123456789ABCDEFGHIJ"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# GitHub credential for GitHub Enterprise Server
resource "seqera_github_credential" "example_github_enterprise" {
  name         = "Example GitHub Enterprise Credentials"
  access_token = "ghp_ExamplePersonalAccessToken123456789ABCDEFGHIJ"
  base_url     = "https://github.example.com"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "github_credential_id" {
  value       = seqera_github_credential.example_github_com.credentials_id
  description = "The ID of the GitHub credential"
}

output "github_credential_provider_type" {
  value       = seqera_github_credential.example_github_com.provider_type
  description = "The provider type (should be 'github')"
}
