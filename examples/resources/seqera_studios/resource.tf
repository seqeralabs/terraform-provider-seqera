resource "seqera_studios" "my_studios" {
  auto_start     = false
  compute_env_id = "compute-env-id"
  configuration = {
    conda_environment = "...my_conda_environment..."
    cpu               = 2
    gpu               = 8
    lifespan_hours    = 2
    memory            = 8192
    mount_data = [
      "..."
    ]
  }
  data_studio_tool_url  = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  description           = "Jupyter studio for data analysis and visualization"
  initial_checkpoint_id = 9
  is_private            = true
  label_ids = [
    7
  ]
  name = "my-jupyter-studio"
  remote_config = {
    commit_id  = "...my_commit_id..."
    repository = "...my_repository..."
    revision   = "...my_revision..."
  }
  spot         = true
  workspace_id = 9
}