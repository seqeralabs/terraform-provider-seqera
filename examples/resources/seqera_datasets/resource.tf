resource "seqera_datasets" "my_datasets" {
  description  = "Dataset containing sample genomic data"
  name         = "my-dataset"
  workspace_id = 7
}