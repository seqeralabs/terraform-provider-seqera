# Azure Cloud with a user-assigned managed identity for VM authentication.
# The credential resource can still use shared keys or Entra service principal;
# the managed identity here is the *runtime* identity attached to each VM.
resource "seqera_azure_cloud_ce" "managed_identity" {
  name           = "azure-cloud-mi"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region                     = "eastus"
    work_dir                   = "az://my-container/work"
    subscription_id            = "00000000-0000-0000-0000-000000000000"
    resource_group             = "my-resource-group"
    instance_type              = "Standard_D4s_v3"
    managed_identity_client_id = "00000000-0000-0000-0000-000000000000"
    managed_identity_id        = "/subscriptions/.../resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/seqera-vm"
  }
}
