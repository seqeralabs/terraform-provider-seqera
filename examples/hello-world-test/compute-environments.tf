resource "seqera_compute_env" "aws_batch_compute_env" {
  workspace_id = resource.seqera_workspace.my_workspace.id
  
  compute_env = {
    name         = "aws-batch-compute-env"
    description  = "AWS Batch compute environment for bioinformatics workflows"
    platform     = "aws-batch"
    credentials_id = resource.seqera_credential.aws_credential.credentials_id
    
    config = {
      aws_batch = {
        discriminator    = "aws-batch"
        region          = "us-east-1" # e.g., "us-east-1"
        work_dir        = local.work_dir # e.g., "s3://my-bucket/work"
        
        
        # Head job configuration
        head_job_cpus      = 2
        head_job_memory_mb = 4096
        
        # Features
        fusion2_enabled = true
        wave_enabled    = true
        

        # Optional: Forge configuration for auto-scaling
        forge = {
          dispose_on_deletion = true
          type               = "EC2" # or "SPOT" for cost savings
          alloc_strategy     = "BEST_FIT_PROGRESSIVE"
          
          # Instance configuration
          instance_types = ["m5.large", "m5.xlarge", "m5.2xlarge"]
          min_cpus      = 0
          max_cpus      = 1000
          

          # Storage
          ebs_auto_scale  = false
          
          # Optional: GPU support
          gpu_enabled = false
          
          # Optional: ARM64 support
          arm64_enabled = false
          
        #   # Optional: EC2 key pair for debugging
        #   ec2_key_pair = var.ec2_key_pair_name
        }
        
        # Optional: Custom Nextflow configuration
        nextflow_config = <<-EOF
          process {
            executor = 'awsbatch'
            queue = 'default'
          }
          aws {
            region = 'us-east-1'}'
            batch {
              cliPath = '/home/ec2-user/miniconda/bin/aws'
            }
          }
        EOF
        
        # Optional: Pre and post-run scripts
        pre_run_script = <<-EOF
          #!/bin/bash
          echo "Starting workflow execution..."
          # Add any setup commands here
        EOF
        
        post_run_script = <<-EOF
          #!/bin/bash
          echo "Workflow execution completed!"
          # Add any cleanup commands here
        EOF
      }
    }
  }
}