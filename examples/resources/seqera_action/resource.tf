# Action Resource Examples
#
# Actions enable event-based pipeline execution in Seqera Platform, such as
# triggering a pipeline launch with a webhook whenever code is updated.
#
# IMPORTANT: GitHub webhook actions require a logged in GitHub account associated with the user creating
# the action.

# Example 1: Basic Tower webhook action (recommended - no GitHub account required)
resource "seqera_action" "tower_basic" {
  workspace_id = seqera_workspace.main.id
  name         = "api-triggered-pipeline"
  source       = "tower"

  config = {
    tower = {
      discriminator = "tower"
    }
  }

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "main"
  }
}

# Example 2: Tower webhook with parameters (using jsonencode and heredoc)
resource "seqera_action" "tower_advanced" {
  workspace_id = seqera_workspace.main.id
  name         = "production-pipeline"
  source       = "tower"

  config = {
    tower = {
      discriminator = "tower"
    }
  }

  launch = {
    pipeline       = "https://github.com/myorg/production-pipeline"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/production/work"
    revision       = "main"

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

# Example 3: GitHub webhook action (requires linked GitHub account)
resource "seqera_action" "github_webhook" {
  workspace_id = seqera_workspace.main.id
  name         = "github-push-trigger"
  source       = "github"

  config = {
    github = {
      discriminator = "github"
    }
  }

  launch = {
    pipeline       = "https://github.com/myorg/my-pipeline"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "main"

    config_profiles = ["docker", "aws"]

    params_text = jsonencode({
      input  = "s3://my-bucket/input.csv"
      output = "s3://my-bucket/results"
    })
  }
}
