# AWS Compute Environment Examples
#
# AWS compute environments define the execution platform where pipelines run.
# Supports AWS Batch, AWS Cloud, and EKS platforms.

# Example 1: Minimal AWS Batch configuration
# Simplest setup with just required fields
resource "seqera_aws_compute_env" "minimal" {
  name           = "aws-batch-minimal"
  workspace_id   = seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-east-1"
    work_dir = "s3://my-nextflow-bucket/work"
  }
}

# Example 2: AWS Batch with Spot instances
# Cost-optimized setup using Spot instances
resource "seqera_aws_compute_env" "spot" {
  name           = "aws-batch-spot"
  workspace_id   = seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-east-1"
    work_dir      = "s3://my-bucket/work"
    enable_fusion = true
    enable_wave   = false # Set explicitly when enable_wave is true

    forge = {
      type                = "SPOT"
      allocation_strategy = "SPOT_CAPACITY_OPTIMIZED"
      bid_percentage      = 70
      min_cpus            = 0
      max_cpus            = 512
      instance_types      = ["m5.xlarge", "m5.2xlarge", "c5.xlarge"]
      ebs_auto_scale      = true
    }
  }
}

# Example 3: Production with VPC, IAM, and EFS
# Full production configuration
resource "seqera_aws_compute_env" "production" {
  name           = "aws-batch-prod"
  workspace_id   = seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id
  description    = "Production AWS Batch environment"

  config = {
    region           = "us-east-1"
    work_dir         = "s3://prod-bucket/work"
    compute_job_role = "arn:aws:iam::123456789012:role/BatchJobRole"
    execution_role   = "arn:aws:iam::123456789012:role/BatchExecutionRole"
    enable_fusion    = true
    enable_wave      = true

    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Loading modules..."
      module load nextflow/23.10.0
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Archiving results..."
      aws s3 sync /tmp/results s3://archive-bucket/
    EOF

    forge = {
      type            = "EC2"
      min_cpus        = 8
      max_cpus        = 512
      instance_types  = ["m5.2xlarge", "m5.4xlarge"]
      ebs_block_size  = 100
      vpc_id          = "vpc-1234567890abcdef0"
      subnets         = ["subnet-12345", "subnet-67890"]
      security_groups = ["sg-12345678"]
      efs_id          = "fs-1234567890abcdef0"
      efs_mount       = "/mnt/efs"
    }
  }
}

# Example 4: GPU-enabled compute environment
# Configured for GPU workloads
resource "seqera_aws_compute_env" "gpu" {
  name           = "aws-batch-gpu"
  workspace_id   = seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-east-1"
    work_dir      = "s3://gpu-bucket/work"
    enable_fusion = true
    enable_wave   = true

    forge = {
      type           = "EC2"
      gpu_enabled    = true
      instance_types = ["p3.2xlarge", "p3.8xlarge"]
      max_cpus       = 256
      ebs_block_size = 200
    }
  }
}

# Example 5: FSx Lustre for high-performance storage
# Uses FSx for Lustre for fast parallel file system
resource "seqera_aws_compute_env" "fsx" {
  name           = "aws-batch-fsx"
  workspace_id   = seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-east-1"
    work_dir      = "s3://my-bucket/work"
    enable_fusion = true
    enable_wave   = true

    forge = {
      type      = "EC2"
      max_cpus  = 512
      fsx_name  = "my-fsx-filesystem"
      fsx_mount = "/fsx"
      fsx_size  = 1200
    }
  }
}

# Example 6: Fargate head job
# Uses Fargate for head job to reduce costs
resource "seqera_aws_compute_env" "fargate_head" {
  name           = "aws-batch-fargate-head"
  workspace_id   = seqera_workspace.main.id
  platform       = "aws-batch"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-east-1"
    work_dir      = "s3://my-bucket/work"
    enable_fusion = true
    enable_wave   = true

    forge = {
      type                 = "EC2"
      max_cpus             = 256
      fargate_head_enabled = true
    }
  }
}
