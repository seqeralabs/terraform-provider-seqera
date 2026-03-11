resource "seqera_workspace" "my_workspace" {
  name        = "my-workspace"
  org_id      = 123456
  full_name   = "my-org/my-workspace"
  visibility  = "PRIVATE"
  description = "Example workspace"
}

resource "seqera_labels" "workspace_label" {
  workspace_id = seqera_workspace.my_workspace.id
  name         = "owner"
  value        = "john-doe"
  resource     = true
  is_default   = false
}
