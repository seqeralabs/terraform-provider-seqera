resource "seqera_workspace_lineage_settings" "my_workspacelineagesettings" {
  credentials_id    = "...my_credentials_id..."
  lineage_default   = false
  org_id            = 3
  provisioning_mode = "AUTO"
  region            = "...my_region..."
  store_name        = "...my_store_name..."
  topic_arn         = "...my_topic_arn..."
  workspace_id      = 1
}