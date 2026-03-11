resource "seqera_action" "tower_advanced" {
  workspace_id = seqera_workspace.main.id
  name         = "production-pipeline"
  source       = "tower"

  launch = {
    pipeline       = "https://github.com/myorg/production-pipeline"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/production/work"
    revision       = "master"

    params_text = jsonencode({
      input_data  = "s3://my-bucket/input/data.csv"
      output_dir  = "s3://my-bucket/results"
      sample_size = 1000
    })

    config_text = <<-EOT
      process {
        executor = 'awsbatch'
        queue    = 'my-production-queue'
        memory   = '8 GB'
        cpus     = 4
      }
    EOT

    pre_run_script = <<-EOT
      #!/bin/bash
      echo "Starting production pipeline run"
      aws s3 sync s3://my-bucket/reference ./reference
    EOT

    post_run_script = <<-EOT
      #!/bin/bash
      echo "Workflow completed"
      aws s3 sync ./results s3://my-bucket/results
    EOT

    resume      = true
    pull_latest = true
  }
}
