resource "seqera_data_studios" "my_datastudios" {
  auto_start            = true
  compute_env_id        = "...my_compute_env_id..."
  conda_environment     = "...my_conda_environment..."
  cpu                   = 2
  data_studio_tool_url  = "...my_data_studio_tool_url..."
  description           = "...my_description..."
  gpu                   = 8
  initial_checkpoint_id = 8
  is_private            = true
  label_ids = [
    8
  ]
  lifespan_hours = 9
  memory         = 5
  mount_data = [
    "..."
  ]
  name         = "...my_name..."
  session_id   = "...my_session_id..."
  spot         = false
  workspace_id = 5
}