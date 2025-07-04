resource "seqera_data_studios" "my_datastudios" {
  auto_start     = true
  compute_env_id = "...my_compute_env_id..."
  configuration = {
    conda_environment = "...my_conda_environment..."
    cpu               = 6
    gpu               = 8
    lifespan_hours    = 5
    memory            = 9
    mount_data = [
      "..."
    ]
  }
  data_studio_tool_url  = "...my_data_studio_tool_url..."
  description           = "...my_description..."
  initial_checkpoint_id = 8
  is_private            = true
  label_ids = [
    8
  ]
  name         = "...my_name..."
  spot         = false
  workspace_id = 5
}