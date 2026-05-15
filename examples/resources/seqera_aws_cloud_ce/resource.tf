# Look up the target organization and workspace by name.
data "seqera_organization" "main" {
  name = "my-organization"
}

data "seqera_workspace" "main" {
  org_id = data.seqera_organization.main.org_id
  name   = "my-workspace"
}

# Minimal AWS Cloud compute environment (Classic mode).
# Seqera picks the worker fleet automatically. Omit `intelligent_compute_config` in this mode.
#
# work_dir notes:
#   - Required and force-new. Changing it replaces the CE.
#   - Seqera Forge implicitly adds the work_dir URI to allow_buckets at
#     CE-create time. If you set allow_buckets explicitly, include the
#     work_dir URI as the trailing entry to match server-side ordering
#     and avoid a forced-replacement diff on subsequent plans.
#   - Pipelines / workflows may override work_dir at launch without
#     mutating the CE — but the override path must be reachable via the
#     CE's allow_buckets / instance-role IAM permissions.
resource "seqera_aws_cloud_ce" "classic" {
  name           = "aws-cloud-classic"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-west-1"
    work_dir = "s3://my-bucket/work"
  }
}
