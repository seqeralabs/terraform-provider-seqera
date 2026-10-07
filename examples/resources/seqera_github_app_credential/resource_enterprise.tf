variable "github_app_id" {
  type = string
}

variable "github_app_client_id" {
  type = string
}

# GitHub Enterprise Server: base_url points at the instance, and installation
# tokens are requested from <base_url>/api/v3. The private key is read from the
# PEM file downloaded from the GitHub App settings page.
resource "seqera_github_app_credential" "enterprise" {
  name         = "github-app-enterprise"
  workspace_id = seqera_workspace.main.id

  app_id      = var.github_app_id
  client_id   = var.github_app_client_id
  private_key = file("${path.module}/github-app.private-key.pem")

  base_url = "https://github.mycompany.com"
}
