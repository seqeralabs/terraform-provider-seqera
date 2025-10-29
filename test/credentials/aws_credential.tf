# AWS Credentials Example
# These are example non-sensitive values for testing

# Basic AWS credential with access key
resource "seqera_aws_credential" "example_basic" {
  name       = "Example AWS Credentials"
  access_key = "AKIAIOSFODNN7EXAMPLE"
  secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# AWS credential with assume role
resource "seqera_aws_credential" "example_with_role" {
  name            = "Example AWS Credentials with AssumeRole"
  access_key      = "AKIAIOSFODNN7EXAMPLE"
  secret_key      = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "aws_credential_id" {
  value       = seqera_aws_credential.example_basic.credentials_id
  description = "The ID of the AWS credential"
}

output "aws_credential_provider_type" {
  value       = seqera_aws_credential.example_basic.provider_type
  description = "The provider type (should be 'aws')"
}
