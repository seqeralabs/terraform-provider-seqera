# Bitbucket Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "bitbucket_username" {
  description = "Bitbucket username"
  type        = string
}

variable "bitbucket_password" {
  description = "Bitbucket app password"
  type        = string
  sensitive   = true
}

# Example: Basic Bitbucket credentials
resource "seqera_bitbucket_credential" "example" {
  name         = "bitbucket-main"
  workspace_id = seqera_workspace.main.id

  username = var.bitbucket_username
  password = var.bitbucket_password
}
