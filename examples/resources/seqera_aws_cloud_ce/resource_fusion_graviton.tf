# AWS Cloud (Classic mode) with Fusion v2, Wave, and Graviton (ARM64).
# Fusion v2 requires Wave; Graviton requires both.
resource "seqera_aws_cloud_ce" "fusion_graviton" {
  name           = "aws-cloud-fusion-graviton"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-cloud"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-west-1"
    work_dir      = "s3://my-bucket/work"
    enable_wave   = true
    enable_fusion = true
    arm64_enabled = true
    instance_type = "m7g.large" # Graviton head node
    ebs_boot_size = 100
  }
}
