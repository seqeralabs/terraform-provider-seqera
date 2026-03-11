resource "seqera_labels" "critical" {
  workspace_id = 123456
  name         = "critical"
  resource     = false
  is_default   = false
}

resource "seqera_labels" "experimental" {
  workspace_id = 123456
  name         = "experimental"
  resource     = false
  is_default   = false
}
