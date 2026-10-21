# `search` filters on the server. Besides free text (matched against the Studio
# name and tool image URL) it understands `keyword:value` filters: `userName`,
# `computeEnvName`, `computeEnvType`, `status`, `label`, `collaborationMode` and
# the `createdAfter`, `createdBefore`, `updatedAfter` and `updatedBefore` dates.
# `status` is one of starting, running, stopping, stopped, errored, building or
# buildFailed. Combine filters by separating them with spaces.
data "seqera_studios_list" "on_compute_env" {
  workspace_id = seqera_workspace.main.id
  search       = "computeEnvName:${seqera_compute_env.main.name}"
}

data "seqera_studios_list" "running" {
  workspace_id = seqera_workspace.main.id
  search       = "status:running"
}

# Filter further in Terraform, for example private Studios built by Wave
locals {
  private_custom_image_studios = [
    for s in data.seqera_studios_list.on_compute_env.studios : s.name
    if s.is_private && s.custom_image
  ]
}

output "studios_on_compute_env" {
  value = {
    total = data.seqera_studios_list.on_compute_env.total_size
    names = [for s in data.seqera_studios_list.on_compute_env.studios : s.name]
  }
}
