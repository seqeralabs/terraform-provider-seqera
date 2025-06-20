resource "seqera_action" "my_action" {
  launch = {
    compute_env_id = "...my_compute_env_id..."
    config_profiles = [
      "..."
    ]
    config_text        = "...my_config_text..."
    date_created       = "2022-02-22T01:05:48.307Z"
    entry_name         = "...my_entry_name..."
    head_job_cpus      = 3
    head_job_memory_mb = 4
    id                 = "...my_id..."
    label_ids = [
      2
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
  source       = "github"
  workspace_id = 3
}