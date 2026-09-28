# AWS Cloud (Classic mode) with Fusion v2 and Graviton (ARM64).
#
# Fusion v2 and Wave are always on for Cloud compute environments — the
# backend enables them and they are not user-settable
resource "seqera_aws_cloud_ce" "fusion_graviton" {
  name           = "aws-cloud-fusion-graviton"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-west-1"
    work_dir      = "s3://my-bucket/work"
    arm64_enabled = true
    instance_type = "m7g.large" # Graviton head node
    ebs_boot_size = 100
  }
}
