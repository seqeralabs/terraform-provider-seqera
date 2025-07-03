data "seqera_workflows" "my_workflows" {
  attributes = [
    "optimized"
  ]
  workflow_id  = "...my_workflow_id..."
  workspace_id = 10
}