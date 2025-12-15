resource "seqera_compute_env" "my_computeenv" {
  compute_env = {
    config = {
      aws_batch = {
        cli_path             = "/home/ec2-user/miniconda/bin/aws"
        compute_job_role     = "arn:aws:iam::123456789012:role/BatchJobRole"
        compute_queue        = "...my_compute_queue..."
        config_type          = "...my_config_type..."
        dragen_instance_type = "...my_dragen_instance_type..."
        dragen_queue         = "...my_dragen_queue..."
        enable_fusion        = false
        enable_wave          = true
        environment = [
          {
            compute = false
            head    = true
            name    = "...my_name..."
            value   = "...my_value..."
          }
        ]
        execution_role = "arn:aws:iam::123456789012:role/BatchExecutionRole"
        forge = {
          alloc_strategy = "SPOT_CAPACITY_OPTIMIZED"
          allow_buckets = [
            "..."
          ]
          arm64_enabled        = true
          bid_percentage       = 20
          dispose_on_deletion  = true
          dragen_ami_id        = "...my_dragen_ami_id..."
          dragen_enabled       = true
          dragen_instance_type = "...my_dragen_instance_type..."
          ebs_auto_scale       = true
          ebs_block_size       = 100
          ebs_boot_size        = 0
          ec2_key_pair         = "my-keypair"
          ecs_config           = "...my_ecs_config..."
          efs_create           = false
          efs_id               = "fs-1234567890abcdef0"
          efs_mount            = "/mnt/efs"
          fargate_head_enabled = false
          fsx_mount            = "/fsx"
          fsx_name             = "my-fsx-filesystem"
          fsx_size             = 1200
          gpu_enabled          = false
          image_id             = "...my_image_id..."
          instance_types = [
            "m5.xlarge",
            "m5.2xlarge",
            "m5.xlarge",
            "m5.2xlarge",
          ]
          max_cpus = 256
          min_cpus = 0
          security_groups = [
            "sg-12345678",
            "sg-12345678",
          ]
          subnets = [
            "subnet-12345",
            "subnet-67890",
            "subnet-12345",
            "subnet-67890",
          ]
          type   = "SPOT"
          vpc_id = "vpc-1234567890abcdef0"
        }
        fusion_snapshots     = true
        head_job_cpus        = 4
        head_job_memory_mb   = 8192
        head_job_role        = "arn:aws:iam::123456789012:role/BatchHeadJobRole"
        head_queue           = "...my_head_queue..."
        log_group            = "...my_log_group..."
        lustre_id            = "...my_lustre_id..."
        nextflow_config      = "...my_nextflow_config..."
        nvme_storage_enabled = true
        post_run_script      = "...my_post_run_script..."
        pre_run_script       = "...my_pre_run_script..."
        region               = "us-east-1"
        storage_type         = "...my_storage_type..."
        volumes = [
          "..."
        ]
        work_dir = "...my_work_dir..."
      }
      seqeracompute_platform = {
        cli_path             = "/home/ec2-user/miniconda/bin/aws"
        compute_job_role     = "arn:aws:iam::123456789012:role/BatchJobRole"
        compute_queue        = "...my_compute_queue..."
        config_type          = "...my_config_type..."
        dragen_instance_type = "...my_dragen_instance_type..."
        dragen_queue         = "...my_dragen_queue..."
        enable_fusion        = true
        enable_wave          = false
        environment = [
          {
            compute = false
            head    = false
            name    = "...my_name..."
            value   = "...my_value..."
          }
        ]
        execution_role = "arn:aws:iam::123456789012:role/BatchExecutionRole"
        forge = {
          alloc_strategy = "SPOT_CAPACITY_OPTIMIZED"
          allow_buckets = [
            "..."
          ]
          arm64_enabled        = true
          bid_percentage       = 20
          dispose_on_deletion  = true
          dragen_ami_id        = "...my_dragen_ami_id..."
          dragen_enabled       = false
          dragen_instance_type = "...my_dragen_instance_type..."
          ebs_auto_scale       = true
          ebs_block_size       = 100
          ebs_boot_size        = 5
          ec2_key_pair         = "my-keypair"
          ecs_config           = "...my_ecs_config..."
          efs_create           = false
          efs_id               = "fs-1234567890abcdef0"
          efs_mount            = "/mnt/efs"
          fargate_head_enabled = false
          fsx_mount            = "/fsx"
          fsx_name             = "my-fsx-filesystem"
          fsx_size             = 1200
          gpu_enabled          = false
          image_id             = "...my_image_id..."
          instance_types = [
            "m5.xlarge",
            "m5.2xlarge",
            "m5.xlarge",
            "m5.2xlarge",
          ]
          max_cpus = 256
          min_cpus = 0
          security_groups = [
            "sg-12345678",
            "sg-12345678",
          ]
          subnets = [
            "subnet-12345",
            "subnet-67890",
            "subnet-12345",
            "subnet-67890",
          ]
          type   = "SPOT"
          vpc_id = "vpc-1234567890abcdef0"
        }
        fusion_snapshots     = true
        head_job_cpus        = 4
        head_job_memory_mb   = 8192
        head_job_role        = "arn:aws:iam::123456789012:role/BatchHeadJobRole"
        head_queue           = "...my_head_queue..."
        log_group            = "...my_log_group..."
        lustre_id            = "...my_lustre_id..."
        nextflow_config      = "...my_nextflow_config..."
        nvme_storage_enabled = true
        post_run_script      = "...my_post_run_script..."
        pre_run_script       = "...my_pre_run_script..."
        region               = "us-east-1"
        storage_type         = "...my_storage_type..."
        volumes = [
          "..."
        ]
        work_dir = "...my_work_dir..."
      }
    }
    credentials_id = "...my_credentials_id..."
    description    = "...my_description..."
    message        = "...my_message..."
    name           = "...my_name..."
    platform       = "google-lifesciences"
  }
  label_ids = [
    6
  ]
  workspace_id = 1
}