resource "seqera_workspace_participant" "user_by_member_id" {
  org_id       = seqera_orgs.main.org_id
  workspace_id = seqera_workspace.main.id
  member_id    = seqera_organization_member.user.member_id
  role         = "maintain"
}
