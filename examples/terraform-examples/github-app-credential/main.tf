terraform {
  required_version = ">= 1.11"

  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
  }
}

provider "seqera" {
  server_url = var.seqera_server_url
}

resource "seqera_github_app_credential" "this" {
  name         = var.credential_name
  workspace_id = var.workspace_id

  app_id      = var.github_app_id
  client_id   = var.github_app_client_id
  private_key = file(var.github_app_private_key_path)

  base_url = var.github_base_url
}

output "github_app_credential_id" {
  description = "Seqera credential ID for the GitHub App."
  value       = seqera_github_app_credential.this.credentials_id
}
