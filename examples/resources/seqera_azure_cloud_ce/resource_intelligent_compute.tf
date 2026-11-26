# Seqera Intelligent Compute (Preview) distributes tasks across multiple Azure
# VMs with optimised scheduling.
resource "seqera_azure_cloud_ce" "intelligent" {
  name           = "azure-cloud-intelligent"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region          = "eastus"
    work_dir        = "az://my-container/work"
    subscription_id = "00000000-0000-0000-0000-000000000000"
    instance_type   = "Standard_D4s_v5"

    intelligent_compute_enabled = true
    intelligent_compute_config = {
      provisioning_model = "spotFirst" # spot | spotFirst | ondemand

      # Restrict the VM sizes the scheduler may use. Omit to let it pick the
      # most cost-optimal sizes per task. Each entry is a full VM size name,
      # or a glob with `*`. A bare `Standard` matches every size.
      machine_types = [
        "Standard_D4s_v5", # exactly this size
        "Standard_E*s_v5", # memory-optimised Esv5 sizes and variants such as Edsv5
      ]
    }
  }
}
