resource "seqera_teams" "data_team" {
  org_id = seqera_orgs.example_org.org_id
  name   = "data-team"
}

resource "seqera_workspace_participant" "team_access" {
  org_id       = seqera_orgs.example_org.org_id
  workspace_id = seqera_workspace.example_workspace.id
  team_id      = seqera_teams.data_team.team_id
  role         = "admin"
}
