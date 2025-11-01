# GitHub Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "github_username" {
  description = "GitHub username"
  type        = string
}

variable "github_access_token" {
  description = "GitHub Personal Access Token"
  type        = string
  sensitive   = true
}

# Example 1: Basic GitHub credentials (GitHub.com)
resource "seqera_github_credential" "example" {
  name         = "github-main"
  workspace_id = seqera_workspace.main.id

  username     = var.github_username
  access_token = var.github_access_token
}

# Example 2: GitHub Enterprise Server
resource "seqera_github_credential" "enterprise" {
  name         = "github-enterprise"
  workspace_id = seqera_workspace.main.id

  username     = var.github_username
  access_token = var.github_access_token
  base_url     = "https://github.mycompany.com"
}
