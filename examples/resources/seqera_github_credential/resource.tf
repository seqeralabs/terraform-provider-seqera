variable "github_username" {
  type = string
}

variable "github_access_token" {
  type      = string
  sensitive = true
}

resource "seqera_github_credential" "example" {
  name         = "github-main"
  workspace_id = seqera_workspace.main.id

  username     = var.github_username
  access_token = var.github_access_token
}

resource "seqera_github_credential" "enterprise" {
  name         = "github-enterprise"
  workspace_id = seqera_workspace.main.id

  username     = var.github_username
  access_token = var.github_access_token
  base_url     = "https://github.mycompany.com"
}
