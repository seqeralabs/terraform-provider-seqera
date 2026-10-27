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
output "studio_revision" {
  description = "Revision Platform resolved (the default branch when revision is unset)"
  value       = seqera_studios.from_git.remote_config.revision
}

output "studio_commit_id" {
  description = "Commit SHA the Studio was created from"
  value       = seqera_studios.from_git.remote_config.commit_id
}

# Image attributes are read-only. For a Git-based Studio they show what Wave
# built from the repository.
output "studio_base_image" {
  value = seqera_studios.from_git.base_image
}

output "studio_custom_image" {
  description = "True when the Studio runs a custom image built by Wave"
  value       = seqera_studios.from_git.custom_image
}

output "studio_wave_build_url" {
  description = "Wave build for the Studio image, if one was built"
  value       = seqera_studios.from_git.wave_build_url
}
