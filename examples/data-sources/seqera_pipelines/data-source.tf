data "seqera_pipelines" "my_pipelines" {
  attributes = [
    "optimized"
  ]
  max          = 3
  offset       = 6
  search       = "...my_search..."
  visibility   = "...my_visibility..."
  workspace_id = 1
}