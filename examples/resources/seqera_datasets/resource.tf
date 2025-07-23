resource "seqera_datasets" "my_datasets" {
  description  = "Research dataset containing sample genomic data"
  name         = "my-research-dataset"
  workspace_id = 7
}
