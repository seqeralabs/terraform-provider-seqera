resource "seqera_aws_batch_compute_env" "my_awsbatchcomputeenv" {
  config = {
    cli_path         = "...my_cli_path..."
    compute_job_role = "...my_compute_job_role..."
    compute_queue    = "...my_compute_queue..."
    enable_fusion    = false
    enable_wave      = false
    execution_role   = "...my_execution_role..."
    forge = {
      allocation_strategy  = "BEST_FIT"
      bid_percentage       = 28
      dispose_on_deletion  = false
      ebs_auto_scale       = false
      ebs_block_size       = 14017
      ec2_key_pair         = "...my_ec2_key_pair..."
      efs_create           = false
      efs_id               = "...my_efs_id..."
      efs_mount            = "...my_efs_mount..."
      fargate_head_enabled = false
      forge_type           = "EC2"
      fsx_mount            = "...my_fsx_mount..."
      fsx_name             = "...my_fsx_name..."
      fsx_size             = 6
      gpu_enabled          = false
      instance_types = [
        "..."
      ]
      max_cpus = 10
      min_cpus = 0
      security_groups = [
        "..."
      ]
      subnets = [
        "..."
      ]
      vpc_id = "...my_vpc_id..."
    }
    head_job_cpus      = 2
    head_job_memory_mb = 9
    head_job_role      = "...my_head_job_role..."
    head_queue         = "...my_head_queue..."
    post_run_script    = "...my_post_run_script..."
    pre_run_script     = "...my_pre_run_script..."
  }
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  label_ids = [
    3
  ]
  name           = "...my_name..."
  region         = "...my_region..."
  work_directory = "...my_work_directory..."
  workspace_id   = 10
}