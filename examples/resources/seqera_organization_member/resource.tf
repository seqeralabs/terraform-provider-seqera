resource "seqera_organization_member" "my_organizationmember" {
  member_id = 8
  org_id    = 1
  role      = "owner"
  user      = "user@example.com"
}