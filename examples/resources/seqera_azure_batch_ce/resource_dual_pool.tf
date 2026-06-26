# Azure Batch with separate head and worker pools.
# Useful when head jobs have different VM-size requirements than workers
# (e.g. memory-heavy head node, cheap auto-scaling workers).
resource "seqera_azure_batch_ce" "dual_pool" {
  name           = "azure-batch-dual-pool"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region   = "eastus"
    work_dir = "az://my-container/work"
    forge = {
      dual_pool_config    = true
      dispose_on_deletion = true
      head_pool = {
        vm_type           = "Standard_E8s_v3"
        vm_count          = 1
        auto_scale        = false
        boot_disk_size_gb = 200
      }
      worker_pool = {
        vm_type           = "Standard_D4s_v3"
        vm_count          = 10
        auto_scale        = true
        boot_disk_size_gb = 100
      }
    }
  }
}
