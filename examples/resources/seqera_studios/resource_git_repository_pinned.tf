# Build a Studio from a Git repository, pinned to a commit, with some of the
# repository's settings overridden.
#
# `revision` accepts a branch, a tag or a full 40-character commit SHA. Branches
# and tags are resolved when the Studio is created, so the Studio keeps the
# commit it was created from even if the branch later moves. A full SHA pins the
# Studio explicitly. `remote_config.commit_id` is read-only and reports the
# resolved commit.
#
# Changing `remote_config` forces the Studio to be replaced.
resource "seqera_studios" "from_git_pinned" {
  name           = "studio-from-git-pinned"
  description    = "Studio built from a pinned commit"
  compute_env_id = seqera_compute_env.main.id
  workspace_id   = seqera_workspace.main.id
  auto_start     = false

  remote_config = {
    repository = "https://github.com/my-org/my-studio"
    revision   = "3f2b8c1d9e4a7b6c5d0e1f2a3b4c5d6e7f8a9b0c"
  }

  # Anything set here takes precedence over `.seqera/studio-config.yaml`.
  # Anything left unset (cpu, memory, lifespan_hours, ...) is taken from the
  # repository, then from the compute environment or workspace defaults.
  is_private = true

  configuration = {
    cpu            = 4
    memory         = 16384
    lifespan_hours = 12
    environment = {
      STUDIO_ENV = "production"
    }
  }

  label_ids = [seqera_labels.team_datascience.id]
}
