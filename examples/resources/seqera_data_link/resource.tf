resource "seqera_data_link" "my_datalink" {
  credentials_id    = "aws-credentials-id"
  description       = "S3 bucket for production data storage"
  name              = "my-s3-datalink"
  provider_type     = "seqeracompute"
  public_accessible = false
  resource_ref      = "s3://my-bucket"
  type              = "bucket"
  workspace_id      = 4
}