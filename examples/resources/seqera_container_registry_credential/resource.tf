variable "registry_username" {
  type = string
}

variable "registry_password" {
  type      = string
  sensitive = true
}

resource "seqera_container_registry_credential" "dockerhub" {
  name         = "dockerhub-main"
  workspace_id = seqera_workspace.main.id

  registry = "docker.io"
  username = var.registry_username
  password = var.registry_password
}

resource "seqera_container_registry_credential" "private" {
  name         = "private-registry"
  workspace_id = seqera_workspace.main.id

  registry = "registry.mycompany.com"
  username = var.registry_username
  password = var.registry_password
}
