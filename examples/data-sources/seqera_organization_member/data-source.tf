data "seqera_organization_member" "my_organizationmember" {
  max    = 10
  offset = 0
  org_id = 1
  search = "...my_search..."
}