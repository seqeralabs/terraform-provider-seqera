data "seqera_dataset" "my_dataset" {
  attributes = [
    "labels"
  ]
  dataset_id   = "...my_dataset_id..."
  workspace_id = 2
}