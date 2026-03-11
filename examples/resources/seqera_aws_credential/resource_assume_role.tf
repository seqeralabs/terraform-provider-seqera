resource "seqera_aws_credential" "with_assume_role" {
  name         = "aws-with-role"
  workspace_id = seqera_workspace.main.id

  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"
}
