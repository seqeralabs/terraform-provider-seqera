# Azure Batch with Fusion v2, Wave, and Fusion metrics collection.
# Fusion v2 mounts blob containers as a distributed file system and
# requires Wave; metrics collection in turn requires Fusion.
resource "seqera_azure_batch_ce" "fusion" {
  name           = "azure-batch-fusion"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  # Fusion metrics collection requires Fusion to be enabled below.
  # Unlike most compute environment settings, this updates in place —
  # flipping it does not replace the CE.
  fusion_metrics_collection_enabled = true

  config = {
    region        = "eastus"
    work_dir      = "az://my-container/work"
    enable_wave   = true
    enable_fusion = true
    forge = {
      vm_type             = "Standard_D4s_v3"
      vm_count            = 5
      auto_scale          = true
      dispose_on_deletion = true
      boot_disk_size_gb   = 100
    }
  }
}
