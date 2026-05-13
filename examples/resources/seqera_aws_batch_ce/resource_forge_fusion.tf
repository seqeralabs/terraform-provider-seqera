# AWS Batch with Forge, Fusion v2, and Wave — the recommended modern setup.
# Forge auto-provisions VPC subnets and security groups; Fusion v2 requires Wave.
#
# work_dir + allow_buckets interaction (worth knowing):
#   - work_dir is the CE's default scratch path; Seqera auto-appends its
#     full URI to the END of allow_buckets on the server side.
#   - Mirror that ordering in config (work_dir URI last) — otherwise plan
#     sees a list-position diff and forces CE replacement on every apply.
#   - work_dir is force-new: changing it replaces the whole CE.
#   - Pipelines / workflows may override work_dir per launch without
#     mutating the CE's allow_buckets — but the overriding bucket must
#     already be in allow_buckets, or the run will fail at runtime.
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
      # Trailing entry MUST match work_dir above — Seqera appends it
      # there itself; mirror it to keep terraform plan a no-op.
      allow_buckets = [
        "s3://my-bucket-input",
        "s3://my-bucket-ref",
        "s3://my-bucket/work",
      ]
    }
  }
}
