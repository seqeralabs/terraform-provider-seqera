resource "seqera_azure_cloud_ce" "my_azurecloudce" {
  config = {
    data_collection_endpoint = "...my_data_collection_endpoint..."
    data_collection_rule_id  = "...my_data_collection_rule_id..."
    enable_fusion            = true
    enable_wave              = false
    environment = [
      {
        compute = false
        head    = false
        name    = "...my_name..."
        value   = "...my_value..."
      }
    ]
    instance_type              = "Standard_D4s_v3"
    log_table_name             = "...my_log_table_name..."
    log_workspace_id           = "...my_log_workspace_id..."
    managed_identity_client_id = "...my_managed_identity_client_id..."
    managed_identity_id        = "...my_managed_identity_id..."
    network_id                 = "...my_network_id..."
    nextflow_config            = "...my_nextflow_config..."
    post_run_script            = "...my_post_run_script..."
    pre_run_script             = "...my_pre_run_script..."
    region                     = "eastus"
    resource_group             = "my-resource-group"
    subscription_id            = "00000000-0000-0000-0000-000000000000"
    work_dir                   = "az://my-container/work"
  }
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  label_ids = [
    6
  ]
  name         = "...my_name..."
  platform     = "azure-cloud"
  workspace_id = 5
}