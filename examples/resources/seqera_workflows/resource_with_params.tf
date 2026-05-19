# Launch a parameterised pipeline run with a workspace-scoped secret.
resource "seqera_workflows" "with_params" {
  workspace_id   = seqera_workspace.main.id
  compute_env_id = seqera_aws_batch_ce.production.compute_env_id
  work_dir       = seqera_aws_batch_ce.production.config.work_dir

  pipeline = "https://github.com/nf-core/rnaseq"
  revision = "3.14.0"
  run_name = "rnaseq-${formatdate("YYYYMMDD-hhmm", timestamp())}"

  params_text = jsonencode({
    input  = "s3://my-bucket/samplesheet.csv"
    outdir = "s3://my-bucket/results"
    genome = "GRCh38"
  })

  workspace_secrets = [
    seqera_pipeline_secret.api_token.name,
  ]
}
