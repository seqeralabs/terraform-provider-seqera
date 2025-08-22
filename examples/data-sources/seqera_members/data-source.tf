data "seqera_members" "my_members" {
  max     = 2
  offset  = 9
  org_id  = 6
  search  = "...my_search..."
  team_id = 3
}