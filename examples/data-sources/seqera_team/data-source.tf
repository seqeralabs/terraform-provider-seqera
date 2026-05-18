# Look up an organization team by name. Useful for resolving the
# `team_id` argument expected by `seqera_workspace_participant`.
data "seqera_organization" "main" {
  name = "my-organization"
}

data "seqera_team" "engineering" {
  org_id = data.seqera_organization.main.org_id
  name   = "engineering"
}

# Assign the team to a workspace as a participant.
resource "seqera_workspace_participant" "engineering_maintain" {
  org_id       = data.seqera_organization.main.org_id
  workspace_id = seqera_workspace.production.id
  team_id      = data.seqera_team.engineering.team_id
  role         = "maintain"
}
