# Seqera Primary Compute Environment Examples
#
# The primary compute environment is the default compute environment used for
# launching workflows within a workspace. Setting a primary compute environment
# allows workflows to run without explicitly specifying a compute environment.

# Example 1: Basic primary compute environment
# Set an existing compute environment as primary

resource "seqera_primary_compute_env" "default" {
  workspace_id   = 123
  compute_env_id = "abc123def456"
}

# Example 2: Primary compute environment with workspace and compute env dependencies
# Shows the full resource relationship

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
