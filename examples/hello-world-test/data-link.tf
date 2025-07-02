# resource "seqera_data_link" "my_datalink" {
#   credentials_id    = resource.seqera_credential.aws_credential.credentials_id
#   description       = "data link created by Terraform"
#   name              = "terraform-datalink"
#   provider_type     = "aws"
#   public_accessible = false
#   type              = "bucket"
#   workspace_id      = resource.seqera_workspace.my_workspace.id
#   resource_ref     = local.work_dir
# }