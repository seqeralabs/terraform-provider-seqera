resource "seqera_labels" "my_labels" {
  is_default   = false
  name         = "environment"
  resource     = true
  value        = "production"
  workspace_id = 1
}
