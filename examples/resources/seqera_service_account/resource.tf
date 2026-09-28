# A service account is an organization-level identity for agents.
resource "seqera_service_account" "ci" {
  org_id      = var.org_id
  name        = "ci-agent"
  description = "Identity for agents triggered by CI pipelines"
}

# Add it to a workspace so agents there can run as it. The role must allow
# launching agents: `launch`, `maintain`, `admin` and `owner` do, while
# `connect` and `view` (the default for service accounts) do not.
resource "seqera_workspace_participant" "ci" {
  org_id       = var.org_id
  workspace_id = var.workspace_id
  member_id    = seqera_service_account.ci.member_id
  role         = "launch"
}
