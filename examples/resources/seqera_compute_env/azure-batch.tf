# Example: Azure Batch Compute Environment
#
# IMPORTANT: Azure Batch work directories must use the az:// protocol format.
# Do not use HTTPS URLs like https://account.blob.core.windows.net/container
#
# Correct format:   az://container-name/path
# Incorrect format: https://account.blob.core.windows.net/container/path
#
# Example conversion:
# HTTPS URL: https://myaccount.blob.core.windows.net/terraform-provider/work
# Correct:   az://terraform-provider/work

resource "seqera_compute_env" "azure_batch_example" {
  compute_env = {
    name           = "azure-batch-environment"
    description    = "Azure Batch compute environment"
    platform       = "azure-batch"
    credentials_id = var.azure_credentials_id

    config = {
      azure_batch = {
        # REQUIRED: Azure region
        region = "eastus"

        # REQUIRED: Work directory using az:// protocol
        # This must be an Azure Blob container, not an HTTPS URL
        work_dir = "az://my-container/work"

        # Optional: Head pool for workflow orchestration
        head_pool = "my-head-pool"

        # Optional: Token duration for Azure authentication
        token_duration = "PT12H"

        # Optional: Job cleanup policy
        delete_jobs_on_completion = "on_success"

        # Optional: Pool cleanup
        delete_pools_on_completion = true

        # Optional: Environment variables
        environment = [
          {
            name    = "AZURE_STORAGE_ACCOUNT"
            value   = "myaccount"
            compute = true
            head    = true
          }
        ]

        # Optional: Enable Wave containers
        wave_enabled = true

        # Optional: Enable Fusion file system
        fusion2_enabled = true

        # Optional: Managed identity for Azure resources
        managed_identity_client_id = var.managed_identity_client_id

        # Optional: Nextflow configuration
        nextflow_config = "process.executor = 'azurebatch'"

        # Optional: TowerForge auto-provisioning
        forge = {
          vm_type             = "Standard_D4s_v3"
          vm_count            = 5
          auto_scale          = true
          dispose_on_deletion = true
          container_reg_ids = [
            "/subscriptions/xxx/resourceGroups/xxx/providers/Microsoft.ContainerRegistry/registries/myregistry"
          ]
        }
      }
    }
  }

  workspace_id = var.workspace_id
}

# Variables
variable "workspace_id" {
  description = "Seqera workspace ID"
  type        = number
}

variable "azure_credentials_id" {
  description = "Azure credentials ID in Seqera"
  type        = string
}

variable "managed_identity_client_id" {
  description = "Optional: Azure managed identity client ID"
  type        = string
  default     = null
}

# Outputs
output "compute_env_id" {
  description = "The ID of the created Azure Batch compute environment"
  value       = seqera_compute_env.azure_batch_example.compute_env_id
}
