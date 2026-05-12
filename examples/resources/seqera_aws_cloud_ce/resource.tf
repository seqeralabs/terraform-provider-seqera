# Minimal AWS Cloud compute environment (Classic mode).
# Seqera picks the worker fleet automatically. Omit `intelligent_compute_config` in this mode.
resource "seqera_aws_cloud_ce" "classic" {
  name           = "aws-cloud-classic"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-cloud"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-west-1"
    work_dir = "s3://my-bucket/work"
  }
}
