variable "gitea_username" {
  type = string
}

variable "gitea_password" {
  type      = string
  sensitive = true
}

resource "seqera_gitea_credential" "example" {
  name         = "gitea-main"
  workspace_id = seqera_workspace.main.id

  username = var.gitea_username
  password = var.gitea_password
  base_url = "https://gitea.mycompany.com"
}
