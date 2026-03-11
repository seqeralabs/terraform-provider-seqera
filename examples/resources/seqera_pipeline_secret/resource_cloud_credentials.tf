variable "aws_access_key" {
  description = "AWS access key for S3 access"
  type        = string
  sensitive   = true
}

resource "seqera_pipeline_secret" "aws_access_key" {
  name         = "aws_access_key_id"
  value        = var.aws_access_key
  workspace_id = seqera_workspace.main.id
}
