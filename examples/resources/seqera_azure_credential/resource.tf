# Azure Batch credential — shared key authentication.
#
# This resource always sets `provider = "azure"` on the wire. For Entra
# or Cloud service-principal credentials (Azure Batch manual mode or
# Azure Cloud SingleVM CE), use the generic `seqera_credential` resource
# with `provider_type = "azure_entra"` or `provider_type = "azure-cloud"`
# until dedicated typed resources land.
resource "seqera_azure_credential" "shared_key" {
  name         = "azure-shared-key"
  workspace_id = seqera_workspace.main.id

  batch_name   = var.azure_batch_name
  batch_key    = var.azure_batch_key
  storage_name = var.azure_storage_name
  storage_key  = var.azure_storage_key
}
