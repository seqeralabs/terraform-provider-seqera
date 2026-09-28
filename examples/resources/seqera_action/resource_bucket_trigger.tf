# Launches a run when the marker file appears in the watched data link.
resource "seqera_action" "bucket_trigger" {
  workspace_id = seqera_workspace.main.id
  name         = "run-on-upload"
  source       = "bucket"

  bucket = {
    data_link_id = seqera_data_link.inputs.data_link_id
    marker_file  = "ready.txt"
  }

  launch = {
    pipeline       = "https://github.com/nf-core/rnaseq"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"

    params_text = jsonencode({
      input  = "s3://my-bucket/inputs/samplesheet.csv"
      outdir = "s3://my-bucket/results"
    })
  }
}
