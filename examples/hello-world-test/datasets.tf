resource "seqera_datasets" "my_datasets" {
  description  = "Terraform created dataset"
  name         = "terraform-dataset"
  workspace_id = resource.seqera_workspace.my_workspace.id
}