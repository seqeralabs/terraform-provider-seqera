# Azure Batch using a user-assigned managed identity instead of shared keys.
# Recommended for production — avoids long-lived Azure access keys in Platform.
resource "seqera_azure_batch_ce" "managed_identity" {
  name           = "azure-batch-mi"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.entra.credentials_id

  config = {
    region                            = "eastus"
    work_dir                          = "az://my-container/work"
    managed_identity_client_id        = "00000000-0000-0000-0000-000000000000"
    managed_identity_head_resource_id = "/subscriptions/.../resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/seqera-head"
    managed_identity_pool_client_id   = "11111111-1111-1111-1111-111111111111"
    managed_identity_pool_resource_id = "/subscriptions/.../resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/seqera-pool"
    forge = {
      vm_type    = "Standard_D4s_v3"
      vm_count   = 5
      auto_scale = true
    }
  }
}
