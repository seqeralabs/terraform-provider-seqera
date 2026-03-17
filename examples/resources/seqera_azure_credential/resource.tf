# Azure credential with shared key authentication
resource "seqera_azure_credential" "shared_key" {
  name         = "azure-shared-key"
  workspace_id = seqera_workspace.main.id

  batch_name   = var.azure_batch_name
  batch_key    = var.azure_batch_key
  storage_name = var.azure_storage_name
  storage_key  = var.azure_storage_key
}

# Azure credential with Entra ID / Cloud authentication (service principal)
resource "seqera_azure_credential" "entra" {
  name         = "azure-entra"
  workspace_id = seqera_workspace.main.id

  batch_name    = var.azure_batch_name
  storage_name  = var.azure_storage_name
  tenant_id     = var.azure_tenant_id
  client_id     = var.azure_client_id
  client_secret = var.azure_client_secret
}
