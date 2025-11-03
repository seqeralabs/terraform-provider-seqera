variable "k8s_certificate" {
  type      = string
  sensitive = true
}

variable "k8s_private_key" {
  type      = string
  sensitive = true
}

resource "seqera_kubernetes_credential" "example" {
  name         = "k8s-main"
  workspace_id = seqera_workspace.main.id

  certificate = var.k8s_certificate
  private_key = var.k8s_private_key
}
