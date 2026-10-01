variable "aws_access_key_id" {
  type      = string
  sensitive = true
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}

resource "seqera_aws_credential" "with_keys_role_and_external_id" {
  name         = "aws-with-keys-role-and-external-id"
  workspace_id = seqera_workspace.main.id

  access_key      = var.aws_access_key_id
  secret_key      = var.aws_secret_access_key
  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraExecutionRole"
  use_external_id = true
}
