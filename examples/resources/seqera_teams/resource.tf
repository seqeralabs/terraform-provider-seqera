resource "seqera_teams" "my_teams" {
  avatar_id    = "avatar-123456"
  description  = "Team responsible for bioinformatics analysis and pipeline development"
  idp_group_id = 2
  name         = "bioinformatics-team"
  org_id       = 1
}