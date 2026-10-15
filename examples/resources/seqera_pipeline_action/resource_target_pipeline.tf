# Links the action to a Launchpad pipeline so each run is recorded as a run of
# that pipeline. The run itself still uses the settings in launch.
resource "seqera_pipeline_action" "linked" {
  workspace_id = seqera_workspace.main.id
  name         = "linked-nightly-run"
  source       = "cron"

  cron = {
    expression = "0 2 * * *"
    timezone   = "UTC"
  }

  pipeline = {
    target_pipeline_id = seqera_pipeline.hello.pipeline_id
  }

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
