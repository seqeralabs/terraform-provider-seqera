variable "ssh_private_key" {
  type      = string
  sensitive = true
}

variable "ssh_passphrase" {
  type      = string
  sensitive = true
  default   = ""
}

resource "seqera_ssh_credential" "example" {
  name         = "ssh-main"
  workspace_id = seqera_workspace.main.id

  private_key = var.ssh_private_key
  passphrase  = var.ssh_passphrase
}
