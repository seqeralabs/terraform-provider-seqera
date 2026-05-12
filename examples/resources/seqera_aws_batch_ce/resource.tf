# Minimal AWS Batch compute environment with Batch Forge.
# Seqera provisions and tears down the underlying Batch infra on apply/destroy.
resource "seqera_aws_batch_ce" "minimal" {
  name           = "aws-batch-minimal"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-east-1"
    work_dir = "s3://my-bucket/work"
    forge = {
      type           = "SPOT"
      max_cpus       = 256
      alloc_strategy = "SPOT_CAPACITY_OPTIMIZED"
    }
  }
}
