resource "seqera_workflows" "my_workflows" {
  compute_env_id = "...my_compute_env_id..."
  config_profiles = [
    "..."
  ]
  config_text        = "...my_config_text..."
  date_created       = "2022-09-13T13:18:33.649Z"
  entry_name         = "...my_entry_name..."
  force              = false
  head_job_cpus      = 0
  head_job_memory_mb = 1
  label_ids = [
    6
  ]
  launch_container     = "...my_launch_container..."
  main_script          = "...my_main_script..."
  optimization_id      = "...my_optimization_id..."
  optimization_targets = "...my_optimization_targets..."
  params_text          = "...my_params_text..."
  pipeline             = "...my_pipeline..."
  post_run_script      = "...my_post_run_script..."
  pre_run_script       = "...my_pre_run_script..."
  pull_latest          = false
  resume               = true
  revision             = "...my_revision..."
  run_name             = "...my_run_name..."
  schema_name          = "...my_schema_name..."
  session_id           = "...my_session_id..."
  source_workspace_id  = 2
  stub_run             = true
  tower_config         = "...my_tower_config..."
  user_secrets = [
    "..."
  ]
  work_dir     = "...my_work_dir..."
  workspace_id = 10
  workspace_secrets = [
    "..."
  ]
}