# Azure Batch credentials — shared key authentication
# provider_type "azure" pairs with keys.azure (Azure Batch CE, Batch Forge).
resource "seqera_credential" "azure_shared_keys" {
  name          = "azure-shared-keys"
  workspace_id  = seqera_workspace.main.id
  provider_type = "azure"
  description   = "Azure Batch credentials using shared keys"
  keys = {
    azure = {
      batch_name   = var.azure_batch_name
      batch_key    = var.azure_batch_key
      storage_name = var.azure_storage_name
      storage_key  = var.azure_storage_key
    }
  }
}

# Azure Batch credentials — Entra service principal
# provider_type "azure_entra" pairs with keys.azure_entra (Azure Batch CE, manual).
resource "seqera_credential" "azure_entra" {
  name          = "azure-entra-sp"
  workspace_id  = seqera_workspace.main.id
  provider_type = "azure_entra"
  description   = "Azure Batch credentials using Entra service principal"
  keys = {
    azure_entra = {
      batch_name    = var.azure_batch_name
      storage_name  = var.azure_storage_name
      tenant_id     = var.azure_tenant_id
      client_id     = var.azure_client_id
      client_secret = var.azure_client_secret
    }
  }
}

# Azure Cloud credentials — Entra service principal
# provider_type "azure-cloud" pairs with keys.azure_cloud (Azure Cloud / SingleVM CE).
resource "seqera_credential" "azure_cloud" {
  name          = "azure-cloud-sp"
  workspace_id  = seqera_workspace.main.id
  provider_type = "azure-cloud"
  description   = "Azure Cloud credentials using Entra service principal"
  keys = {
    azure_cloud = {
      subscription_id = var.azure_subscription_id
      storage_name    = var.azure_storage_name
      tenant_id       = var.azure_tenant_id
      client_id       = var.azure_client_id
      client_secret   = var.azure_client_secret
    }
  }
}
