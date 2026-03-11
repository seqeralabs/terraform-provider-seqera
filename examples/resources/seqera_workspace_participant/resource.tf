resource "seqera_orgs" "example_org" {
  name      = "example-org"
  full_name = "Example Organization"
}

resource "seqera_workspace" "example_workspace" {
  org_id     = seqera_orgs.example_org.org_id
  name       = "example-workspace"
  full_name  = "Example Workspace"
  visibility = "PRIVATE"
}

resource "seqera_workspace_participant" "user_by_email" {
  org_id       = seqera_orgs.example_org.org_id
  workspace_id = seqera_workspace.example_workspace.id
  email        = "user@example.com"
  role         = "launch"
}
