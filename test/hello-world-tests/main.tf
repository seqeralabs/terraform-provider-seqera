## Organizations
resource "seqera_orgs" "test_org" {
  description = "testing org for the terraform provider"
  full_name   = "seqera_test_shahbaz_tf_provider"
  name        = "seqera_test_shahbaz_tf_provider"
}

## Workspaces
resource "seqera_workspace" "my_workspace" {
  description = "A test workspace created with Terraform"
  name        = "test-workspace-tf"
  full_name   = "Test Workspace for Terraform Provider"
  org_id      = resource.seqera_orgs.test_org.org_id
  visibility  = "PRIVATE"
}

## Teams
resource "seqera_teams" "my_teams" {
  description = "Team created by Terraform"
  name        = "terraform-test-team"
  org_id      = resource.seqera_orgs.test_org.org_id
}


module "aws_batch" {
  source             = "./modules/aws"
  iam_role           = "arn:aws:iam::128997144437:role/TowerDevelopmentRole"
  work_dir           = "s3://shahbaz-test"
  access_key         = var.access_key
  secret_key         = var.secret_key
  workspace_id       = resource.seqera_workspace.my_workspace.id
  seqera_bearer_auth = var.seqera_bearer_auth
  seqera_server_url  = var.seqera_server_url
}

# module "gcp_batch" {
#   source              = "./modules/gcp"
#   work_dir            = local.gcp_work_dir
#   workspace_id        = resource.seqera_workspace.my_workspace.id
#   service_account_key = local.service_account_key
#   seqera_bearer_auth  = var.seqera_bearer_auth
#   seqera_server_url   = var.seqera_server_url
# }

module "azure_batch" {
  source             = "./modules/azure"
  batch_key          = var.azure_batch_key
  batch_name         = local.azure_batch_name
  storage_key        = var.azure_storage_key
  storage_name       = local.azure_storage_name
  seqera_bearer_auth = var.seqera_bearer_auth
  seqera_server_url  = var.seqera_server_url
  workspace_id       = resource.seqera_workspace.my_workspace.id
  work_dir           = local.azure_work_dir
}
