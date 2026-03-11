resource "seqera_action" "tower_basic" {
  workspace_id = seqera_workspace.main.id
  name         = "api-triggered-pipeline"
  source       = "tower"

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
