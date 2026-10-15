# Moves a pipeline action managed by seqera_action to seqera_pipeline_action
# without recreating it. Requires Terraform 1.8 or later. Replace the
# seqera_action block with a seqera_pipeline_action one with the same
# arguments, dropping response_type if it was set.
moved {
  from = seqera_action.nightly
  to   = seqera_pipeline_action.nightly
}

resource "seqera_pipeline_action" "nightly" {
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
