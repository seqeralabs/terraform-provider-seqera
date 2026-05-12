# AWS Batch with Forge, Fusion v2, and Wave — the recommended modern setup.
# Forge auto-provisions VPC subnets and security groups; Fusion v2 requires Wave.
resource "seqera_aws_batch_ce" "forge_fusion" {
  name           = "aws-batch-forge-fusion"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-east-1"
    work_dir      = "s3://my-bucket/work"
    enable_wave   = true
    enable_fusion = true
    forge = {
      type                = "SPOT"
      max_cpus            = 512
      alloc_strategy      = "SPOT_CAPACITY_OPTIMIZED"
      dispose_on_deletion = true
      ebs_boot_size       = 100
      allow_buckets       = ["s3://my-bucket-input", "s3://my-bucket-ref"]
    }
  }
}
