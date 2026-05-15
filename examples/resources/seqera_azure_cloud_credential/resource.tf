# Azure Cloud credential — Entra service principal authentication.
# Use this credential type for Azure Cloud (SingleVM) compute environments.
resource "seqera_azure_cloud_credential" "cloud" {
  name         = "azure-cloud-sp"
  workspace_id = seqera_workspace.main.id

  subscription_id = var.azure_subscription_id
  storage_name    = var.azure_storage_name
  tenant_id       = var.azure_tenant_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
}
