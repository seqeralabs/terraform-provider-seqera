# Look up the target organization and workspace by name.
data "seqera_organization" "main" {
  name = "my-organization"
}

data "seqera_workspace" "main" {
  org_id = data.seqera_organization.main.org_id
  name   = "my-workspace"
}

# Minimal AWS Cloud compute environment (Classic mode).
# Seqera picks the worker fleet automatically.
#
# If you set `allow_buckets` explicitly, include the `work_dir` URI as the
# trailing entry — Seqera Forge implicitly appends it at CE-create time, and
# omitting it produces a forced-replacement diff on subsequent plans.
resource "seqera_aws_cloud_ce" "classic" {
  name           = "aws-cloud-classic"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-west-1"
    work_dir = "s3://my-bucket/work"
  }
}
