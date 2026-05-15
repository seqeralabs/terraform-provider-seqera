# Minimal Azure Cloud compute environment.
# Nextflow runs directly on Azure VMs managed by Seqera (no Batch service).
resource "seqera_azure_cloud_ce" "minimal" {
  name           = "azure-cloud-minimal"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region          = "eastus"
    work_dir        = "az://my-container/work"
    subscription_id = "00000000-0000-0000-0000-000000000000"
    instance_type   = "Standard_D4s_v3"
    # resource_group is read-only — Forge creates a `TowerForge-<ce>-<id>`
    # resource group per compute environment and exposes its name on the
    # state attribute after apply.
  }
}
