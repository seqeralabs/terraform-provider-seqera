# lifespan_hours = 0 removes the limit on Studio session length. This is also
# the only setting that lets each Studio choose its own
# `configuration.lifespan_hours`.
#
# With any other value, Platform only accepts Studios in the workspace whose
# configuration.lifespan_hours equals the workspace value (see the next example).
resource "seqera_workspace_studios_settings" "unlimited" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  lifespan_hours            = 0
  private_studio_by_default = true

  container_repository = "cr.seqera.io/my-org/studios"
  name_strategy        = "imageSuffix"
}

resource "seqera_studios" "custom_lifespan" {
  name                 = "studio-custom-lifespan"
  compute_env_id       = seqera_compute_env.main.id
  workspace_id         = seqera_workspace.main.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"

  configuration = {
    lifespan_hours = 24
  }

  # The settings must exist first so the limit is in place when the Studio is
  # created.
  depends_on = [seqera_workspace_studios_settings.unlimited]
}
