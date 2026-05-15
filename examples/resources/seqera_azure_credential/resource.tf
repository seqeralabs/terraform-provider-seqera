# Azure Batch credential — shared key authentication.
resource "seqera_azure_credential" "shared_key" {
  name         = "azure-shared-key"
  workspace_id = seqera_workspace.main.id

  batch_name   = var.azure_batch_name
  batch_key    = var.azure_batch_key
  storage_name = var.azure_storage_name
  storage_key  = var.azure_storage_key
}
