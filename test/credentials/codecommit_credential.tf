# AWS CodeCommit Credentials Example
# These are example non-sensitive values for testing

# CodeCommit credential
resource "seqera_codecommit_credential" "example" {
  name       = "Example CodeCommit Credentials"
  access_key = "AKIAIOSFODNN7EXAMPLE"
  secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  base_url   = "https://git-codecommit.us-east-1.amazonaws.com/v1/repos/example-repo"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "codecommit_credential_id" {
  value       = seqera_codecommit_credential.example.credentials_id
  description = "The ID of the CodeCommit credential"
}

output "codecommit_credential_provider_type" {
  value       = seqera_codecommit_credential.example.provider_type
  description = "The provider type (should be 'codecommit')"
}
