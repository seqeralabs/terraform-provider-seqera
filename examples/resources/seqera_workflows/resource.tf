# seqera_workflows — launch an individual workflow run on the Seqera
# Platform. Each resource maps to one Run on the Runs page.
#
# Use this when Terraform itself should trigger a run (e.g. a validation
# launch after a compute environment change, or a CI fan-out across CEs).
# For a saved, re-launchable pipeline definition on the Launchpad, use
# `seqera_pipeline` instead.

# Minimal launch against an existing compute environment.
resource "seqera_workflows" "hello" {
  workspace_id   = seqera_workspace.main.id
  compute_env_id = seqera_aws_batch_ce.production.compute_env_id
  work_dir       = seqera_aws_batch_ce.production.config.work_dir

  pipeline = "https://github.com/nextflow-io/hello"
  revision = "master"
  run_name = "hello-tf-launch"
}
