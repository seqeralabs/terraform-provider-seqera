variable "codecommit_username" {
  type = string
}

variable "codecommit_password" {
  type      = string
  sensitive = true
}

resource "seqera_codecommit_credential" "example" {
  name         = "codecommit-main"
  workspace_id = seqera_workspace.main.id

  username = var.codecommit_username
  password = var.codecommit_password
}
