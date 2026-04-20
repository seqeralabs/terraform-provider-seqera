resource "seqera_aws_credential" "with_external_id" {
  name         = "aws-with-external-id"
  workspace_id = seqera_workspace.main.id

  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"
  mode            = "role"
  use_external_id = true
}

# Use this external ID in your AWS IAM role trust policy:
output "seqera_external_id" {
  description = "External ID for the AWS IAM role trust policy"
  value       = seqera_aws_credential.with_external_id.external_id
}
