resource "seqera_workspace_participant" "my_workspaceparticipant" {
  member_id          = 1
  participant_id     = 4
  role               = "connect"
  team_id            = 2
  user_name_or_email = "user@example.com"
  workspace_id       = 9
}