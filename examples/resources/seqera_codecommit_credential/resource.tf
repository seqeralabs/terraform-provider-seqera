# AWS CodeCommit Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "codecommit_username" {
  description = "CodeCommit username"
  type        = string
}

variable "codecommit_password" {
  description = "CodeCommit password"
  type        = string
  sensitive   = true
}

# Example: Basic CodeCommit credentials
resource "seqera_codecommit_credential" "example" {
  name         = "codecommit-main"
  workspace_id = seqera_workspace.main.id

  username = var.codecommit_username
  password = var.codecommit_password
}
