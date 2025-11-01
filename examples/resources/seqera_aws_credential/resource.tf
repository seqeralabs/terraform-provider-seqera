# AWS Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "aws_access_key_id" {
  description = "AWS Access Key ID"
  type        = string
  sensitive   = true
}

variable "aws_secret_access_key" {
  description = "AWS Secret Access Key"
  type        = string
  sensitive   = true
}

# Example: Basic AWS credentials
resource "seqera_aws_credential" "example" {
  name         = "aws-main"
  workspace_id = seqera_workspace.main.id

  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}
