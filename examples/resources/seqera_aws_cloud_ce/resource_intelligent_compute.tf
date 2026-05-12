# Seqera Intelligent Compute distributes tasks across multiple EC2 instances
# with optimised scheduling. See the resource docs for prerequisites and
# the SEQERA_SCHEDULER feature toggle requirement.
resource "seqera_aws_cloud_ce" "intelligent" {
  name           = "aws-cloud-intelligent"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-cloud"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-west-1"
    work_dir      = "s3://my-bucket/work"
    allow_buckets = ["s3://my-bucket-input", "s3://my-bucket-ref"]
    sched_enabled = true
    sched_config = {
      provisioning_model = "spotFirst" # spot | spotFirst | ondemand
      machine_types      = []          # empty = scheduler picks cost-optimal
    }
  }
}
