resource "seqera_teams" "my_teams" {
  description = "Team created by Terraform"
  name        = "terraform-test-team"
  org_id      = resource.seqera_orgs.test_org.org_id
}


