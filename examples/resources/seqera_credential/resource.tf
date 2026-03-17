# Azure Batch credentials (shared key authentication)
resource "seqera_credential" "azure_batch" {
  name          = "azure-batch-creds"
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

# Azure Cloud credentials (service principal / Entra ID)
resource "seqera_credential" "azure_cloud" {
  name          = "azure-cloud-creds"
  workspace_id  = seqera_workspace.main.id
  provider_type = "azure"
  description   = "Azure Cloud credentials using service principal"
  keys = {
    azure_cloud = {
      tenant_id       = var.azure_tenant_id
      client_id       = var.azure_client_id
      client_secret   = var.azure_client_secret
      subscription_id = var.azure_subscription_id
    }
  }
}

# Azure Entra ID credentials
resource "seqera_credential" "azure_entra" {
  name          = "azure-entra-creds"
  workspace_id  = seqera_workspace.main.id
  provider_type = "azure_entra"
  description   = "Azure Entra ID credentials"
  keys = {
    azure_entra = {
      tenant_id     = var.azure_tenant_id
      client_id     = var.azure_client_id
      client_secret = var.azure_client_secret
      batch_name    = var.azure_batch_name
      storage_name  = var.azure_storage_name
    }
  }
}
