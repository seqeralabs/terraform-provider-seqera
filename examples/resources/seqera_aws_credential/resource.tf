resource "seqera_aws_credential" "my_awscredential" {
  base_url       = "...my_base_url..."
  category       = "...my_category..."
  checked        = false
  credentials_id = "...my_credentials_id..."
  date_created   = "2022-09-21T21:03:12.536Z"
  deleted        = true
  description    = "...my_description..."
  keys = {
    access_key      = "AKIAIOSFODNN7EXAMPLE"
    assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraRole"
    secret_key      = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  last_updated  = "2022-07-20T00:51:49.763Z"
  last_used     = "2021-06-04T18:43:01.971Z"
  name          = "...my_name..."
  provider_type = "aws"
  workspace_id  = 4
}