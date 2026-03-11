# IMPORTANT: GitHub webhook actions require a logged in GitHub account associated
# with the user creating the action.
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
    revision       = "master"

    config_profiles = ["docker", "aws"]

    params_text = jsonencode({
      input  = "s3://my-bucket/input.csv"
      output = "s3://my-bucket/results"
    })
  }
}
