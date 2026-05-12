# AWS Batch with pre-existing infra — bring your own compute queue and IAM.
# No `forge` block: Seqera neither provisions nor tears down Batch resources.
resource "seqera_aws_batch_ce" "manual" {
  name           = "aws-batch-manual"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region           = "us-east-1"
    work_dir         = "s3://my-bucket/work"
    compute_queue    = "my-existing-batch-queue"
    head_queue       = "my-existing-batch-head-queue"
    head_job_role    = "arn:aws:iam::123456789012:role/SeqeraHeadJobRole"
    compute_job_role = "arn:aws:iam::123456789012:role/SeqeraComputeJobRole"
    execution_role   = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"
  }
}
