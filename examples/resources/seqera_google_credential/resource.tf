variable "gcp_service_account_key" {
  type      = string
  sensitive = true
}

resource "seqera_google_credential" "example" {
  name         = "gcp-main"
  workspace_id = seqera_workspace.main.id

  key = var.gcp_service_account_key
}
