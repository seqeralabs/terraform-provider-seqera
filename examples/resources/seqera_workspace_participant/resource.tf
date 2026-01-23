#
# Create an Organization.
#---
resource "seqera_orgs" "example_org" {
  name      = "example-org"
  full_name = "Example Organization"
}

#
# Create a Workspace.
#---
resource "seqera_workspace" "example_workspace" {
  org_id     = seqera_orgs.example_org.org_id
  name       = "example-workspace"
  full_name  = "Example Workspace"
  visibility = "PRIVATE"
}

#
# Add a user to the organization.
#---
resource "seqera_organization_member" "user" {
  org_id = seqera_orgs.example_org.org_id
  email  = "user@example.com"
}

#
# Create a team.
#---
resource "seqera_teams" "data_team" {
  org_id = seqera_orgs.example_org.org_id
  name   = "data-team"
}

#
# Add the user to the team.
#---
resource "seqera_team_member" "user_in_team" {
  org_id    = seqera_teams.data_team.org_id
  team_id   = seqera_teams.data_team.team_id
  member_id = seqera_organization_member.user.member_id
}

#
# Example 1: Add an individual user to workspace by email.
# Email lookup happens once during creation, then participant_id is cached in state.
#---
resource "seqera_workspace_participant" "user_by_email" {
  org_id       = seqera_orgs.example_org.org_id
  workspace_id = seqera_workspace.example_workspace.id
  email        = "user@example.com"
  role         = "launch"
}

#
# Example 2: Add an individual user to workspace by member_id.
# Uses the member_id directly with no lookup.
#---
resource "seqera_workspace_participant" "user_by_member_id" {
  org_id       = seqera_orgs.example_org.org_id
  workspace_id = seqera_workspace.example_workspace.id
  member_id    = seqera_organization_member.user.member_id
  role         = "maintain"
}

#
# Example 3: Add an entire team to workspace.
# All team members receive the specified role.
#---
resource "seqera_workspace_participant" "team_access" {
  org_id       = seqera_orgs.example_org.org_id
  workspace_id = seqera_workspace.example_workspace.id
  team_id      = seqera_teams.data_team.team_id
  role         = "admin"
}
