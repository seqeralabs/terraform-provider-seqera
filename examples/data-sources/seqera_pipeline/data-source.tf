data "seqera_pipeline" "my_pipeline" {
  attributes = [
    "computeEnv"
  ]
  source_workspace_id = 6
  workspace_id        = 3
}