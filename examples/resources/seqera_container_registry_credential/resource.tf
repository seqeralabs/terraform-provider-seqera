# Container Registry Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "registry_username" {
  description = "Container registry username"
  type        = string
}

variable "registry_password" {
  description = "Container registry password or token"
  type        = string
  sensitive   = true
}

# Example 1: Docker Hub
resource "seqera_container_registry_credential" "dockerhub" {
  name         = "dockerhub-main"
  workspace_id = seqera_workspace.main.id

  registry = "docker.io"
  username = var.registry_username
  password = var.registry_password
}

# Example 2: Private registry
resource "seqera_container_registry_credential" "private" {
  name         = "private-registry"
  workspace_id = seqera_workspace.main.id

  registry = "registry.mycompany.com"
  username = var.registry_username
  password = var.registry_password
}
