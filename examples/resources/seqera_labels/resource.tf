resource "seqera_labels" "my_labels" {
  is_default   = false
  label_id     = 4
  name         = "...my_name..."
  resource     = false
  value        = "...my_value..."
  workspace_id = 1
}