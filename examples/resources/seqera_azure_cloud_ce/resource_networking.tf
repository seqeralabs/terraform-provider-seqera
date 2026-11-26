# Azure Cloud in an existing virtual network instead of one Seqera creates.
# Use this for private connectivity to other Azure resources, or in a region
# without the StandardV2 NAT gateway Seqera uses for the network it creates.
resource "seqera_azure_cloud_ce" "networking" {
  name           = "azure-cloud-vnet"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region          = "eastus"
    work_dir        = "az://my-container/work"
    subscription_id = "00000000-0000-0000-0000-000000000000"
    instance_type   = "Standard_D4s_v5"

    # The virtual network must be in `region`.
    network_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-network-rg/providers/Microsoft.Network/virtualNetworks/my-vnet"

    # Names (not resource IDs) of subnets in that network. Instances launch in
    # the first one. Omit to use a subnet discovered in the network.
    subnets = ["compute"]
  }
}
