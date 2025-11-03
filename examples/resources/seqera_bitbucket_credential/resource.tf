variable "bitbucket_username" {
  type = string
}

variable "bitbucket_password" {
  type      = string
  sensitive = true
}

resource "seqera_bitbucket_credential" "example" {
  name         = "bitbucket-main"
  workspace_id = seqera_workspace.main.id

  username = var.bitbucket_username
  password = var.bitbucket_password
}
