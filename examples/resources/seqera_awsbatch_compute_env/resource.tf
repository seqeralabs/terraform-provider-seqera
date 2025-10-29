# AWS Batch Compute Environment Examples
#
# AWS Batch compute environments provide scalable compute capacity for running
# Nextflow workflows on AWS using the AWS Batch service.

# Example 1: Minimal AWS Batch configuration
# Simplest setup with just required fields
resource "seqera_awsbatch_compute_env" "minimal" {
  name           = "aws-batch-minimal"
  workspace_id   = 123
  credentials_id = "aws-creds-id"
  region         = "us-east-1"
  work_directory = "s3://my-bucket/work"
}

# Example 2: AWS Batch with Spot instances
# Cost-optimized setup using Spot instances with bid percentage
resource "seqera_awsbatch_compute_env" "spot" {
  name           = "aws-batch-spot"
  workspace_id   = 123
  credentials_id = "aws-creds-id"
  region         = "us-east-1"
  work_directory = "s3://my-bucket/work"

  forge_type          = "SPOT"
  allocation_strategy = "SPOT_CAPACITY_OPTIMIZED"
  bid_percentage      = 70 # Pay up to 70% of On-Demand price
  min_cpus            = 0
  max_cpus            = 512
  instance_types      = ["m5.xlarge", "m5.2xlarge", "c5.xlarge", "c5.2xlarge"]
  ebs_auto_scale      = true
  enable_fusion       = true
  enable_wave         = true
}

# Example 3: Production setup with VPC and EFS
# Full production configuration with custom VPC, IAM roles, and EFS storage
resource "seqera_awsbatch_compute_env" "production" {
  name           = "aws-batch-prod"
  workspace_id   = 123
  credentials_id = "aws-creds-id"
  region         = "us-east-1"
  work_directory = "s3://prod-bucket/work"
  description    = "Production AWS Batch environment"

  # Compute configuration
  forge_type     = "EC2"
  min_cpus       = 8
  max_cpus       = 512
  instance_types = ["m5.2xlarge", "m5.4xlarge"]
  ebs_block_size = 100

  # Network configuration
  vpc_id          = "vpc-1234567890abcdef0"
  subnets         = ["subnet-12345", "subnet-67890", "subnet-abcde"]
  security_groups = ["sg-12345678"]

  # IAM roles
  compute_job_role = "arn:aws:iam::123456789012:role/BatchJobRole"
  execution_role   = "arn:aws:iam::123456789012:role/BatchExecutionRole"

  # EFS configuration
  efs_id    = "fs-1234567890abcdef0"
  efs_mount = "/mnt/efs"

  # Fusion and Wave
  enable_fusion = true
  enable_wave   = true

  # Head job configuration
  head_job_cpus      = 4
  head_job_memory_mb = 8192

  # Lifecycle scripts
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
}

# Example 4: GPU-enabled compute environment
# Configured for GPU workloads with p3 instances
resource "seqera_awsbatch_compute_env" "gpu" {
  name           = "aws-batch-gpu"
  workspace_id   = 123
  credentials_id = "aws-creds-id"
  region         = "us-east-1"
  work_directory = "s3://gpu-bucket/work"

  forge_type      = "EC2"
  gpu_enabled     = true
  instance_types  = ["p3.2xlarge", "p3.8xlarge"]
  max_cpus        = 256
  ebs_block_size  = 200
  enable_fusion   = true
  enable_wave     = true
}

# Example 5: Fargate head job configuration
# Uses Fargate for head job to reduce costs
resource "seqera_awsbatch_compute_env" "fargate_head" {
  name           = "aws-batch-fargate-head"
  workspace_id   = 123
  credentials_id = "aws-creds-id"
  region         = "us-east-1"
  work_directory = "s3://my-bucket/work"

  forge_type           = "EC2"
  fargate_head_enabled = true
  max_cpus             = 256
  enable_fusion        = true
  enable_wave          = true
}

# Example 6: FSx Lustre configuration
# High-performance storage with FSx for Lustre
resource "seqera_awsbatch_compute_env" "fsx" {
  name           = "aws-batch-fsx"
  workspace_id   = 123
  credentials_id = "aws-creds-id"
  region         = "us-east-1"
  work_directory = "s3://my-bucket/work"

  forge_type = "EC2"
  max_cpus   = 512

  # FSx configuration
  fsx_name  = "my-fsx-filesystem"
  fsx_mount = "/fsx"
  fsx_size  = 1200

  enable_fusion = true
  enable_wave   = true
}
