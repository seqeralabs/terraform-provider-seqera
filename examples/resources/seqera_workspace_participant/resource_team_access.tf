resource "seqera_workspace_participant" "team_access" {
  org_id       = seqera_orgs.main.org_id
  workspace_id = seqera_workspace.main.id
  team_id      = seqera_teams.data_team.team_id
  role         = "admin"
}
