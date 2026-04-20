# NOTE: role-based credentials without access keys and without a platform-generated
# external ID are only supported on enterprise Seqera Platform installs that have
# the relevant feature flags enabled. On Seqera cloud (api.cloud.seqera.io), use
# the External Id variant (use_external_id = true) or provide access keys.
resource "seqera_aws_credential" "with_assume_role" {
  name         = "aws-with-role"
  workspace_id = seqera_workspace.main.id

  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"
  mode            = "role"
}
