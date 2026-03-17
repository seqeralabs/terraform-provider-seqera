resource "seqera_labels" "critical" {
  workspace_id = seqera_workspace.main.id
  name         = "critical"
  resource     = false
  is_default   = false
}
