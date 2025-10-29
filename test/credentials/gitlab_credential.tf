# GitLab Credentials Example
# These are example non-sensitive values for testing

# GitLab credential for GitLab.com
resource "seqera_gitlab_credential" "example_gitlab_com" {
  name  = "Example GitLab Credentials"
  token = "glpat-ExamplePersonalAccessToken1234567890AB"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# GitLab credential for self-hosted GitLab instance
resource "seqera_gitlab_credential" "example_self_hosted" {
  name     = "Example GitLab Self-Hosted Credentials"
  token    = "glpat-ExamplePersonalAccessToken1234567890AB"
  base_url = "https://gitlab.example.com"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "gitlab_credential_id" {
  value       = seqera_gitlab_credential.example_gitlab_com.credentials_id
  description = "The ID of the GitLab credential"
}

output "gitlab_credential_provider_type" {
  value       = seqera_gitlab_credential.example_gitlab_com.provider_type
  description = "The provider type (should be 'gitlab')"
}
