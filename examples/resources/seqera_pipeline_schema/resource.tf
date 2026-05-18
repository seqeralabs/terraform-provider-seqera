# Pipeline Schema Example
#
# Create a Nextflow pipeline schema and bind it to a pipeline via
# launch.pipeline_schema_id.
#
# The Seqera API exposes pipeline schemas as write-once objects:
#   - POST /pipeline-schemas is the only operation available.
#   - There is no GET, PUT, or DELETE by schema id.
#
# As a result, schema_content changes force resource replacement (the API
# has no update), and destroying this resource leaves the schema row
# orphaned server-side — there's nothing the provider can call to clean it up.

resource "seqera_pipeline_schema" "rnaseq" {
  workspace_id   = seqera_workspace.main.id
  schema_content = file("${path.module}/nextflow_schema.json")
}

resource "seqera_pipeline" "rnaseq" {
  workspace_id = seqera_workspace.main.id
  name         = "rna-seq-analysis"

  launch = {
    pipeline           = "https://github.com/nf-core/rnaseq"
    compute_env_id     = seqera_compute_env.aws.compute_env.id
    work_dir           = "s3://my-bucket/work"
    revision           = "3.14.0"
    pipeline_schema_id = seqera_pipeline_schema.rnaseq.id
  }

  # The pipeline_schema_id reference already implies ordering; depends_on is
  # included explicitly so the relationship is obvious to readers and survives
  # refactors (e.g. if the id is ever passed via a local or variable).
  depends_on = [seqera_pipeline_schema.rnaseq]
}
