# Studios settings for a workspace. Platform stores a single set of settings per
# workspace, so only one of these resources should manage a given workspace.
#
# The Platform API replaces all settings on every change, so lifespan_hours and
# private_studio_by_default are always required. Unset optional fields fall back
# to the platform defaults.
#
# Destroying this resource only removes it from Terraform state: Platform has no
# way to delete the settings, so the workspace keeps its current values.
resource "seqera_workspace_studios_settings" "main" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  # Maximum session lifespan in hours; 0 means unlimited. When this is not 0,
  # Platform only accepts Studios in the workspace whose
  # configuration.lifespan_hours equals this value.
  lifespan_hours = 8

  # New Studios are private to their creator unless is_private says otherwise.
  private_studio_by_default = true

  # Where Wave pushes images it builds for Studios (for example from a
  # remote_config or conda environment), and how those images are named.
  container_repository = "cr.seqera.io/my-org/studios"
  name_strategy        = "tagPrefix" # none | tagPrefix | imageSuffix
}
