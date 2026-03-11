resource "seqera_organization_member" "user" {
  org_id = seqera_orgs.example_org.org_id
  email  = "user@example.com"
}

resource "seqera_workspace_participant" "user_by_member_id" {
  org_id       = seqera_orgs.example_org.org_id
  workspace_id = seqera_workspace.example_workspace.id
  member_id    = seqera_organization_member.user.member_id
  role         = "maintain"
}
