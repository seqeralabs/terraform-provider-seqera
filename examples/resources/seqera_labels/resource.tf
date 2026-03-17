resource "seqera_labels" "environment" {
  workspace_id = seqera_workspace.main.id
  name         = "environment"
  value        = "production"
  resource     = true
  is_default   = false
}
