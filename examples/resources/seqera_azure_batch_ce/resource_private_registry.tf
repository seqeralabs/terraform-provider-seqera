# Azure Batch Forge with credentials for private container registries.
resource "seqera_container_registry_credential" "acr" {
  name         = "private-acr"
  workspace_id = data.seqera_workspace.main.id
  registry     = "myregistry.azurecr.io"
  user_name    = var.acr_username
  password     = var.acr_password
}

resource "seqera_azure_batch_ce" "private_registry" {
  name           = "azure-batch-private-registry"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region   = "eastus"
    work_dir = "az://my-container/work"
    forge = {
      vm_type             = "Standard_D4s_v3"
      vm_count            = 5
      auto_scale          = true
      dispose_on_deletion = true
      container_reg_ids = [
        seqera_container_registry_credential.acr.credentials_id
      ]
    }
  }
}
