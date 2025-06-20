data "seqera_pipeline" "my_pipeline" {
  attributes = [
    "computeEnv"
  ]
  pipeline_id         = 1
  source_workspace_id = 6
  workspace_id        = 3
}