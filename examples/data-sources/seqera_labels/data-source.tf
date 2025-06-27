data "seqera_labels" "my_labels" {
  is_default   = false
  max          = 5
  offset       = 10
  search       = "...my_search..."
  type         = "simple"
  workspace_id = 1
}