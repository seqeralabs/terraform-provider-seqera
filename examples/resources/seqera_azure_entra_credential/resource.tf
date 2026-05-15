# Azure Batch credential — Entra service principal authentication.
# Use this credential type for Azure Batch CEs configured in manual mode.
resource "seqera_azure_entra_credential" "entra" {
  name         = "azure-entra-sp"
  workspace_id = seqera_workspace.main.id

  batch_name    = var.azure_batch_name
  storage_name  = var.azure_storage_name
  tenant_id     = var.azure_tenant_id
  client_id     = var.azure_client_id
  client_secret = var.azure_client_secret
}
