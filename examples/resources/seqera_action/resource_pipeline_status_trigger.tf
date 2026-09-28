# Launches a downstream run each time a run of the watched pipeline succeeds.
# The watched pipeline is set in pipeline_status; the run the action starts is
# set in launch.
resource "seqera_action" "on_upstream_success" {
  workspace_id = seqera_workspace.main.id
  name         = "run-after-upstream"
  source       = "pipeline_status"

  pipeline_status = {
    pipeline_id = seqera_pipeline.upstream.pipeline_id
    run_status  = "SUCCEEDED"
  }

  launch = {
    pipeline       = "https://github.com/myorg/downstream-pipeline"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
