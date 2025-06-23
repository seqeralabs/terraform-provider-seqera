data "seqera_data_studios" "my_datastudios" {
  attributes = [
    "labels"
  ]
  max          = 3
  offset       = 9
  search       = "...my_search..."
  workspace_id = 5
}