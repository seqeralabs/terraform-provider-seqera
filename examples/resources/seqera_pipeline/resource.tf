resource "seqera_pipeline" "my_pipeline" {
  description = "...my_description..."
  icon        = "...my_icon..."
  label_ids = [
    7
  ]
  launch = {
    compute_env_id = "...my_compute_env_id..."
    config_profiles = [
      "..."
    ]
    config_text        = "...my_config_text..."
    date_created       = "2022-01-03T00:33:46.725Z"
    entry_name         = "...my_entry_name..."
    head_job_cpus      = 6
    head_job_memory_mb = 4
    label_ids = [
      3
    ]
    launch_container     = "...my_launch_container..."
    main_script          = "...my_main_script..."
    optimization_id      = "...my_optimization_id..."
    optimization_targets = "...my_optimization_targets..."
    params_text          = "...my_params_text..."
    pipeline             = "...my_pipeline..."
    post_run_script      = "...my_post_run_script..."
    pre_run_script       = "...my_pre_run_script..."
    pull_latest          = true
    resume               = false
    revision             = "...my_revision..."
    run_name             = "...my_run_name..."
    schema_name          = "...my_schema_name..."
    session_id           = "...my_session_id..."
    stub_run             = false
    tower_config         = "...my_tower_config..."
    user_secrets = [
      "..."
    ]
    work_dir = "...my_work_dir..."
    workspace_secrets = [
      "..."
    ]
  }
  name         = "...my_name..."
  workspace_id = 3
}