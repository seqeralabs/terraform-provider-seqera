variable "codecommit_access_key" {
  type = string
}

variable "codecommit_secret_key" {
  type      = string
  sensitive = true
}

resource "seqera_codecommit_credential" "example" {
  name         = "codecommit-main"
  workspace_id = seqera_workspace.main.id

  access_key = var.codecommit_access_key
  secret_key = var.codecommit_secret_key
  base_url   = "https://git-codecommit.eu-west-1.amazonaws.com"
}
