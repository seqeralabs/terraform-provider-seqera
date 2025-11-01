variable "gitlab_username" {
  type = string
}

variable "gitlab_token" {
  type      = string
  sensitive = true
}

resource "seqera_gitlab_credential" "example" {
  name         = "gitlab-main"
  workspace_id = seqera_workspace.main.id

  username = var.gitlab_username
  token    = var.gitlab_token
}

resource "seqera_gitlab_credential" "self_hosted" {
  name         = "gitlab-enterprise"
  workspace_id = seqera_workspace.main.id

  username = var.gitlab_username
  token    = var.gitlab_token
  base_url = "https://gitlab.mycompany.com"
}
