# Pipeline Examples
#
# Manage Nextflow pipeline definitions in Seqera Platform.

# Example 1: Pipeline with explicit work_dir
resource "seqera_pipeline" "basic" {
  workspace_id = seqera_workspace.main.id
  name         = "rna-seq-analysis"
  description  = "RNA sequencing analysis pipeline"

  launch = {
    pipeline       = "https://github.com/nf-core/rnaseq"
    compute_env_id = seqera_compute_env.aws.compute_env.id
    work_dir       = "s3://my-bucket/work"
    revision       = "main"
  }
}

# Example 2: Pipeline referencing work_dir from a generic compute environment
# Avoids duplicating the work_dir value across resources
resource "seqera_pipeline" "from_generic_ce" {
  workspace_id = seqera_workspace.main.id
  name         = "genomics-pipeline"

  launch = {
    pipeline       = "https://github.com/nf-core/sarek"
    compute_env_id = seqera_compute_env.aws.compute_env.id
    work_dir       = seqera_compute_env.aws.compute_env.config.aws_batch.work_dir
    revision       = "main"
  }
}

# Example 3: Pipeline referencing work_dir from a platform-specific compute environment
resource "seqera_pipeline" "from_platform_ce" {
  workspace_id = seqera_workspace.main.id
  name         = "variant-calling"

  launch = {
    pipeline       = "https://github.com/nf-core/sarek"
    compute_env_id = seqera_aws_batch_compute_env.production.compute_env_id
    work_dir       = seqera_aws_batch_compute_env.production.config.work_dir
    revision       = "3.4.0"
  }
}

# Example 4: Pipeline in a shared workspace (work_dir and compute_env_id are optional)
resource "seqera_pipeline" "shared" {
  workspace_id = seqera_workspace.shared.id
  name         = "hello-world"

  launch = {
    pipeline = "https://github.com/nextflow-io/hello"
    revision = "main"
  }
}
