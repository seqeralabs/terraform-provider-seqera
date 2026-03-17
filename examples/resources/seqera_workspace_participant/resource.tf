resource "seqera_workspace_participant" "user_by_email" {
  org_id       = seqera_orgs.main.org_id
  workspace_id = seqera_workspace.main.id
  email        = "user@example.com"
  role         = "launch"
}
