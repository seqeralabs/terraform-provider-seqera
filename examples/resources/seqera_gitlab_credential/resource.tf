# GitLab Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "gitlab_username" {
  description = "GitLab username"
  type        = string
}

variable "gitlab_token" {
  description = "GitLab Personal Access Token or Project Access Token"
  type        = string
  sensitive   = true
}

# Example 1: Basic GitLab credentials (GitLab.com)
resource "seqera_gitlab_credential" "example" {
  name         = "gitlab-main"
  workspace_id = seqera_workspace.main.id

  username = var.gitlab_username
  token    = var.gitlab_token
}

# Example 2: Self-hosted GitLab Server
resource "seqera_gitlab_credential" "self_hosted" {
  name         = "gitlab-enterprise"
  workspace_id = seqera_workspace.main.id

  username = var.gitlab_username
  token    = var.gitlab_token
  base_url = "https://gitlab.mycompany.com"
}
