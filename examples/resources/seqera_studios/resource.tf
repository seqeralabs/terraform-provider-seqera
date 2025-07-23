resource "seqera_studios" "my_studios" {
  auto_start     = false
  compute_env_id = "ce-123456789"
  configuration = {
    conda_environment = "...my_conda_environment..."
    cpu               = 2
    gpu               = 8
    lifespan_hours    = 2
    memory            = 3
    mount_data = [
      "..."
    ]
  }
  data_studio_tool_url  = "https://jupyter.org/try-jupyter/hub/lab"
  description           = "Jupyter studio for data analysis and visualization"
  initial_checkpoint_id = 9
  is_private            = true
  label_ids = [
    7
  ]
  name         = "my-jupyter-studio"
  spot         = true
  workspace_id = 9
}
