# Grant another user access to a private Studio.
#
# The Studio owner is always the identity running Terraform: the studio is
# created under the account that owns the TOWER_ACCESS_TOKEN (or provider
# `bearer_auth`) credential. That owner is always allowed to connect and start
# the Studio and never needs to be listed in `allowed_user_ids` — if you include
# the owner's ID, the platform silently drops it from the list.
#
# `allowed_user_ids` takes Seqera *user IDs*, and is only valid on a private
# studio (is_private = true). Each allowed user must already be a member of the
# Studio's workspace, otherwise the platform rejects the request. Currently the
# allow list is capped at one additional user.

# Look up the user's numeric user ID from their email. `seqera_organization_member`
# exposes `user_id`, which is the value `allowed_user_ids` expects.
# (The `seqera_workspace_participant` data source only returns member/participant
# IDs, not the user ID, so it is not used here.)
data "seqera_organization_member" "collaborator" {
  org_id = seqera_workspace.main.org_id
  email  = "collaborator@example.com"
}

resource "seqera_studios" "shared_private" {
  name                 = "shared-private-studio"
  compute_env_id       = seqera_compute_env.main.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  workspace_id         = seqera_workspace.main.id
  configuration        = {}

  # Private studio: only the owner (the Terraform identity) and the users listed
  # below can connect to and start it.
  is_private       = true
  allowed_user_ids = [data.seqera_organization_member.collaborator.user_id]
}

# The resolved membership is available read-only, including the owner:
output "studio_allowed_users" {
  value = seqera_studios.shared_private.allowed_users
}
