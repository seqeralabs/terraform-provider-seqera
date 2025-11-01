# Kubernetes Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "k8s_certificate" {
  description = "Kubernetes client certificate (base64 encoded)"
  type        = string
  sensitive   = true
}

variable "k8s_private_key" {
  description = "Kubernetes client private key (base64 encoded)"
  type        = string
  sensitive   = true
}

# Example: Basic Kubernetes credentials using certificates
resource "seqera_kubernetes_credential" "example" {
  name         = "k8s-main"
  workspace_id = seqera_workspace.main.id

  certificate = var.k8s_certificate
  private_key = var.k8s_private_key
}
