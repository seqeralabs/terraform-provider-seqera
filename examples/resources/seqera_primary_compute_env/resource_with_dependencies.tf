resource "seqera_workspace" "analysis" {
  name      = "analysis-workspace"
  full_name = "my-org/analysis-workspace"
}

resource "seqera_compute_env" "default" {
  name         = "default-compute-env"
  workspace_id = seqera_workspace.analysis.id

  compute_env = {
    config = {
      # Your compute environment configuration here
      # See seqera_compute_env or seqera_aws_compute_env examples
    }
  }
}

resource "seqera_primary_compute_env" "primary" {
  workspace_id   = seqera_workspace.analysis.id
  compute_env_id = seqera_compute_env.default.id
}
