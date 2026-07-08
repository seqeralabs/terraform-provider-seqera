# Seqera Intelligent Compute distributes tasks across multiple EC2 instances
# with optimised scheduling. See the resource docs for prerequisites and
# the SEQERA_SCHEDULER feature toggle requirement.
resource "seqera_aws_cloud_ce" "intelligent" {
  name           = "aws-cloud-intelligent"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-west-1"
    work_dir      = "s3://my-bucket/work"
    allow_buckets = ["s3://my-bucket-input", "s3://my-bucket-ref"]
    intelligent_compute_enabled = true
    intelligent_compute_config = {
      provisioning_model = "spotFirst" # spot | spotFirst | ondemand
      machine_types      = []          # empty = scheduler picks cost-optimal
      backend_strategy   = "ECS"       # ECS (default) | EC2 | VM
      fusion_snapshots   = true        # resume interrupted tasks from a snapshot
      prediction_model   = "none"      # none | qr/v1 | qr/v2

      # Warm pool: keep idle VMs ready for sub-5s task starts, scaling to zero
      # after 5 minutes of inactivity.
      pool = {
        enabled            = true
        desired_warm       = 1
        scale_to_zero_secs = 300
      }
    }
  }
}
