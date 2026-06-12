variable "github_app_id" {
  type = string
}

variable "github_app_client_id" {
  type = string
}

variable "github_app_private_key" {
  type      = string
  sensitive = true
}

resource "seqera_github_app_credential" "example" {
  name         = "github-app-main"
  workspace_id = seqera_workspace.main.id

  app_id    = var.github_app_id
  client_id = var.github_app_client_id

  private_key = var.github_app_private_key

  base_url = "https://github.com/seqeralabs"
}
