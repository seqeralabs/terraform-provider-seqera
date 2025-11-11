# AWS Credential Examples
#
# AWS credentials can be configured in two ways:
# 1. Using access keys (access_key + secret_key)
# 2. Using IAM role assumption (assume_role_arn)

variable "aws_access_key_id" {
  type      = string
  sensitive = true
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}

# Example 1: Using access keys
resource "seqera_aws_credential" "with_keys" {
  name         = "aws-with-keys"
  workspace_id = seqera_workspace.main.id

  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}

# Example 2: Using IAM role assumption
# The Seqera Platform will use ambient AWS credentials to assume this role
resource "seqera_aws_credential" "with_assume_role" {
  name         = "aws-with-role"
  workspace_id = seqera_workspace.main.id

  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"
}

# Example 3: Using both access keys and role assumption
# The access keys will be used to assume the specified role
resource "seqera_aws_credential" "with_keys_and_role" {
  name         = "aws-with-keys-and-role"
  workspace_id = seqera_workspace.main.id

  access_key      = var.aws_access_key_id
  secret_key      = var.aws_secret_access_key
  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"
}
