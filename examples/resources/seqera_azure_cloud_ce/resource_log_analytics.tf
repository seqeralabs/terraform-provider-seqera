# Azure Cloud with telemetry forwarded to a Log Analytics workspace.
# Use this when you want pipeline logs queryable in Azure Monitor.
resource "seqera_azure_cloud_ce" "log_analytics" {
  name           = "azure-cloud-logs"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id

  config = {
    region                   = "eastus"
    work_dir                 = "az://my-container/work"
    subscription_id          = "00000000-0000-0000-0000-000000000000"
    instance_type            = "Standard_D4s_v3"
    log_workspace_id         = "/subscriptions/.../resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/my-law"
    log_table_name           = "SeqeraNextflowLogs"
    data_collection_endpoint = "https://my-dce-xxxx.eastus-1.ingest.monitor.azure.com"
    data_collection_rule_id  = "/subscriptions/.../resourceGroups/rg/providers/Microsoft.Insights/dataCollectionRules/seqera-dcr"
  }
}
