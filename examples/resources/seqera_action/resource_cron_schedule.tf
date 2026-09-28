# Launches a run every night at 02:00 UTC.
resource "seqera_action" "nightly" {
  workspace_id = seqera_workspace.main.id
  name         = "nightly-run"
  source       = "cron"

  cron = {
    expression = "0 2 * * *"
    timezone   = "UTC"
  }

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
