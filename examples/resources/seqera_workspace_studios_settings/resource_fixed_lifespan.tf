# Enforce one lifespan across the workspace. Because lifespan_hours is not 0,
# Platform rejects Studios whose configuration.lifespan_hours differs, so the
# Studio below reads the value from the settings resource instead of repeating it.
resource "seqera_workspace_studios_settings" "fixed" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  lifespan_hours            = 8
  private_studio_by_default = true
}

resource "seqera_studios" "matches_workspace" {
  name                 = "studio-matching-workspace"
  compute_env_id       = seqera_compute_env.main.id
  workspace_id         = seqera_workspace.main.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"

  configuration = {
    lifespan_hours = seqera_workspace_studios_settings.fixed.lifespan_hours
  }
}
