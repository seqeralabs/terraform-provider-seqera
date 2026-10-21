# List the Studios in a workspace
data "seqera_studios_list" "all" {
  workspace_id = seqera_workspace.main.id
}

# Narrow the result server-side. Free text matches the Studio name and tool
# image URL; `userId`, `computeEnvId` and `status` keywords are also supported.
data "seqera_studios_list" "jupyter" {
  workspace_id = seqera_workspace.main.id
  search       = "jupyter"
}

# The platform returns at most 100 Studios per request. Fail early if the
# result was truncated rather than silently missing Studios.
check "studios_not_truncated" {
  assert {
    condition     = data.seqera_studios_list.all.total_size <= length(data.seqera_studios_list.all.studios)
    error_message = "More Studios exist than were returned; narrow the query with `search`."
  }
}

# Look up a Studio by name
locals {
  studios_by_name = {
    for s in data.seqera_studios_list.all.studios : s.name => s
  }
}

output "analysis_studio_session_id" {
  value = local.studios_by_name["my-jupyter-studio"].session_id
}
