# Minimal Azure Batch compute environment with Forge auto-provisioning.
# Uses shared key authentication from the referenced Azure credential.
resource "seqera_azure_batch_ce" "minimal" {
  name           = "azure-batch-minimal"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "azure-batch"
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region   = "eastus"
    work_dir = "az://my-container/work"
    forge = {
      vm_type             = "Standard_D4s_v3"
      vm_count            = 5
      auto_scale          = true
      dispose_on_deletion = true
    }
  }
}
