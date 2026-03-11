resource "seqera_labels" "environment" {
  workspace_id = 123456
  name         = "environment"
  value        = "production"
  resource     = true
  is_default   = false
}
