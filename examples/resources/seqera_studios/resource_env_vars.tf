resource "seqera_studios" "studio_with_env_vars" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
  configuration = {
    cpu            = 2
    memory         = 8192
    lifespan_hours = 8
    # Studio-specific environment variables (keys must be alphanumeric + underscore, cannot start with number)
    environment = {
      MY_STUDIO_VAR = "testing"
      API_ENDPOINT  = "https://api.example.com"
      DEBUG_MODE    = "true"
    }
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-ride:2025.04.1-0.8"
  description          = "Studio with custom environment variables"
  is_private           = true
  name                 = "studio-with-env"
  workspace_id         = seqera_workspace.main.id
}
