resource "seqera_aws_credential" "my_awscredential" {
  base_url       = "...my_base_url..."
  category       = "...my_category..."
  checked        = false
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  keys = {
    access_key      = "AKIAIOSFODNN7EXAMPLE"
    assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraRole"
    secret_key      = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  name          = "...my_name..."
  provider_type = "aws"
  workspace_id  = 4
}