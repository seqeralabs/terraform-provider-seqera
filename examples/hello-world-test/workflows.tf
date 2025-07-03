# resource "seqera_workflows" "my_workflows" {
#   force = false
#     #compute_env_id = resource.seqera_compute_envs.my_compute_env.id
#     # config_profiles = [
#     #   "..."
#     # ]
#     # config_text        = "...my_config_text..."
#     # date_created       = "2022-06-14T18:52:38.830Z"
#     # entry_name         = "...my_entry_name..."
#     # head_job_cpus      = 4
#     # head_job_memory_mb = 5
#     # label_ids = [
#     #   5
#     # ]
#     # launch_container     = "...my_launch_container..."
#     # main_script          = "...my_main_script..."
#     # optimization_id      = "...my_optimization_id..."
#     # optimization_targets = "...my_optimization_targets..."
#     # params_text          = "...my_params_text..."
#     pipeline             = resource.seqera_pipeline.hello_world_minimal.name
#     #post_run_script      = "...my_post_run_script..."
#     #pre_run_script       = "...my_pre_run_script..."
#     pull_latest          = true
#     resume               = false
#     revision             = "master"
#     run_name             = "terraform-test-run"
#     #schema_name          = "...my_schema_name..."
#     #session_id           = "...my_session_id..."
#     #stub_run             = false
#     #tower_config         = "...my_tower_config..."
#     # user_secrets = [
#     #   "..."
#     # ]
#     work_dir = local.work_dir
#     # workspace_secrets = [
#     #   "..."
#     # ]
  
#   #source_workspace_id = resource.seqera_workspace.my_workspace.id
#   workspace_id        = resource.seqera_workspace.my_workspace.id
# }

resource "seqera_workflows" "my_workflows" {
  compute_env_id = resource.seqera_compute_env.aws_batch_compute_env.compute_env_id
  pipeline             = resource.seqera_pipeline.hello_world_minimal.name
  pull_latest          = true
  resume               = true
  revision             = "master"
  source_workspace_id  = resource.seqera_workspace.my_workspace.id
  stub_run             = true
  work_dir     = local.work_dir
  workspace_id = resource.seqera_workspace.my_workspace.id
}