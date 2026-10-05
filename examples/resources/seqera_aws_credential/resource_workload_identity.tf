# The IAM role must trust this Seqera installation's OIDC provider. Build its
# trust policy from the seqera_aws_credentials_federation_setup data source;
# see the "AWS Credentials with Workload Identity Federation" guide.
resource "seqera_aws_credential" "wif" {
  name         = "aws-wif"
  workspace_id = seqera_workspace.main.id

  mode            = "workloadIdentity"
  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraWorkloadIdentityRole"
}
