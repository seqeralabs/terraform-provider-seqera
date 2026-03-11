resource "seqera_studios" "basic_jupyter" {
  name                 = "my-jupyter-studio"
  compute_env_id       = "compute-env-id"
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  workspace_id         = seqera_workspace.my_workspace.id
  # Configuration is required - gpu defaults to 0
  configuration = {}
}
