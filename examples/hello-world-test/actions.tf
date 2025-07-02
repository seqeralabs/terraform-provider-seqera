resource "seqera_action" "my_action" {
  launch = {
    compute_env_id = resource.seqera_compute_env.aws_batch_compute_env.compute_env_id
    config_profiles = [ ]
    config_text        = ""
    pipeline             = "https://github.com/nextflow-io/hello"
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
      
    pull_latest          = true
    resume               = true
    revision             = "master"
    #run_name             = "...my_run_name..." this should be auto generated
    work_dir = local.work_dir
  }
  name         = "terraform-hello-world-action"
  workspace_id = resource.seqera_workspace.my_workspace.id
  source = "tower" 
}