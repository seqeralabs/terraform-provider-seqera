# Build a Studio from a Git repository.
#
# Platform reads `.seqera/studio-config.yaml` from the given revision and builds
# the image with Wave on first start, so `data_studio_tool_url` must not be set:
# exactly one of `data_studio_tool_url` or `remote_config` is required.
#
# Anything left unset (cpu, memory, lifespan_hours, ...) is taken from the
# repository's Studio configuration, then from the compute environment or
# workspace defaults. Values set here take precedence over the repository.
resource "seqera_studios" "from_git" {
  name           = "studio-from-git"
  compute_env_id = seqera_compute_env.main.id
  workspace_id   = seqera_workspace.main.id
  auto_start     = false

  remote_config = {
    repository = "https://github.com/my-org/my-studio"
    revision   = "main" # optional; defaults to the repository's default branch
  }

  configuration = {}
}

# Resolved by Platform and available after creation
output "studio_commit_id" {
  value = seqera_studios.from_git.remote_config.commit_id
}

output "studio_base_image" {
  value = seqera_studios.from_git.base_image
}
