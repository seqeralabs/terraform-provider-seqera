## Organizations
resource "seqera_orgs" "test_org" {
  description = "testing org for the terraform provider - GCP"
  full_name   = "seqera_test_shahbaz_tf_provider_gcp"
  name        = "seqera_test_shahbaz_tf_provider_gcp"
}

## Workspaces
resource  "seqera_workspace" "my_workspace" {
  description = "A test workspace created with Terraform for GCP"
  name        = "test-workspace-tf-gcp"
  full_name   = "Test Workspace for Terraform Provider"
  org_id     = resource.seqera_orgs.test_org.org_id
  visibility = "PRIVATE"
}

## Datasets
resource "seqera_datasets" "my_datasets" {
  description  = "Terraform created dataset"
  name         = "terraform-dataset"
  workspace_id = resource.seqera_workspace.my_workspace.id
}