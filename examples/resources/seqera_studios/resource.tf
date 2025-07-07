resource "seqera_studios" "my_studios" {
  auto_start     = false
  compute_env_id = "...my_compute_env_id..."
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
  data_studio_tool_url  = "...my_data_studio_tool_url..."
  description           = "...my_description..."
  initial_checkpoint_id = 9
  is_private            = true
  label_ids = [
    7
  ]
  name         = "...my_name..."
  spot         = true
  workspace_id = 9
}