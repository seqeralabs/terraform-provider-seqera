resource "seqera_aws_compute_env" "my_awscomputeenv" {
  config = {
    cli_path             = "...my_cli_path..."
    compute_job_role     = "...my_compute_job_role..."
    compute_queue        = "...my_compute_queue..."
    dragen_instance_type = "...my_dragen_instance_type..."
    dragen_queue         = "...my_dragen_queue..."
    environment = [
      {
        compute = false
        head    = true
        name    = "...my_name..."
        value   = "...my_value..."
      }
    ]
    execution_role = "...my_execution_role..."
    forge = {
      alloc_strategy = "BEST_FIT_PROGRESSIVE"
      allow_buckets = [
        "..."
      ]
      arm64_enabled        = true
      bid_percentage       = 8
      dispose_on_deletion  = false
      dragen_ami_id        = "...my_dragen_ami_id..."
      dragen_enabled       = true
      dragen_instance_type = "...my_dragen_instance_type..."
      ebs_auto_scale       = false
      ebs_block_size       = 2
      ebs_boot_size        = 3
      ec2_key_pair         = "...my_ec2_key_pair..."
      ecs_config           = "...my_ecs_config..."
      efs_create           = true
      efs_id               = "...my_efs_id..."
      efs_mount            = "...my_efs_mount..."
      fargate_head_enabled = true
      fsx_mount            = "...my_fsx_mount..."
      fsx_name             = "...my_fsx_name..."
      fsx_size             = 8
      fusion_enabled       = true
      gpu_enabled          = true
      image_id             = "...my_image_id..."
      instance_types = [
        "..."
      ]
      max_cpus = 4
      min_cpus = 1
      security_groups = [
        "..."
      ]
      subnets = [
        "..."
      ]
      type   = "EC2"
      vpc_id = "...my_vpc_id..."
    }
    fusion_snapshots      = false
    fusion2_enabled       = true
    head_job_cpus         = 9
    head_job_memory_mb    = 3
    head_job_role         = "...my_head_job_role..."
    head_queue            = "...my_head_queue..."
    log_group             = "...my_log_group..."
    lustre_id             = "...my_lustre_id..."
    nextflow_config       = "...my_nextflow_config..."
    nvnme_storage_enabled = true
    post_run_script       = "...my_post_run_script..."
    pre_run_script        = "...my_pre_run_script..."
    region                = "...my_region..."
    storage_type          = "...my_storage_type..."
    volumes = [
      "..."
    ]
    wave_enabled = false
    work_dir     = "...my_work_dir..."
  }
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  label_ids = [
    1
  ]
  name         = "...my_name..."
  platform     = "aws-batch"
  workspace_id = 7
}