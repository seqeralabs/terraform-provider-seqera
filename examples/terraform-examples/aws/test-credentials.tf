# Test AWS Credentials with validation
# This tests the flattened structure with nested keys

terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
  }
}

variable "aws_access_key" {
  type        = string
  sensitive   = true
  description = "AWS Access Key ID for testing"
}

variable "aws_secret_key" {
  type        = string
  sensitive   = true
  description = "AWS Secret Access Key for testing"
}

variable "workspace_id" {
  type        = number
  description = "Workspace ID for testing"
}

# Test basic credentials with validation
resource "seqera_aws_credential" "test_basic" {
  name         = "test-aws-basic"
  workspace_id = var.workspace_id

  keys = {
    access_key = var.aws_access_key
    secret_key = var.aws_secret_key
  }
}

# Test with assume role ARN
resource "seqera_aws_credential" "test_with_role" {
  name         = "test-aws-with-role"
  workspace_id = var.workspace_id

  keys = {
    access_key      = var.aws_access_key
    secret_key      = var.aws_secret_key
    assume_role_arn = "arn:aws:iam::123456789012:role/TestSeqeraRole"
  }
}

output "basic_credential_id" {
  value       = seqera_aws_credential.test_basic.credentials_id
  description = "Created credential ID"
}

output "role_credential_id" {
  value       = seqera_aws_credential.test_with_role.credentials_id
  description = "Created credential ID with role"
}
