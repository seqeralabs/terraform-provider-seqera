# Seqera Managed Compute with environment variables, pre/post scripts,
# and a Nextflow config snippet. Useful for workspace-wide defaults.
resource "seqera_managed_compute_ce" "with_env" {
  name          = "seqera-cloud-with-env"
  workspace_id  = data.seqera_workspace.main.id
  region        = "us-east-1"
  instance_size = "MEDIUM"

  environment = [
    {
      name    = "NXF_OPTS"
      value   = "-Xms256m -Xmx2g"
      head    = true
      compute = false
    },
    {
      name    = "MY_API_TOKEN"
      value   = "set-via-env-or-secret-store"
      head    = true
      compute = true
    },
  ]

  pre_run_script  = "echo 'pipeline starting'"
  post_run_script = "echo 'pipeline complete'"
  nextflow_config = "process.cpus = 2"
}
