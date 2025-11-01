# Gitea Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "gitea_username" {
  description = "Gitea username"
  type        = string
}

variable "gitea_password" {
  description = "Gitea password or access token"
  type        = string
  sensitive   = true
}

# Example: Basic Gitea credentials
resource "seqera_gitea_credential" "example" {
  name         = "gitea-main"
  workspace_id = seqera_workspace.main.id

  username = var.gitea_username
  password = var.gitea_password
  base_url = "https://gitea.mycompany.com"
}
