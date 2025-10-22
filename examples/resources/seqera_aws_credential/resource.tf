resource "seqera_aws_credential" "my_awscredential" {
  base_url       = "...my_base_url..."
  category       = "...my_category..."
  checked        = false
  credentials_id = "...my_credentials_id..."
  description    = "AWS credentials for Seqera workflow execution"
  keys = {
    access_key      = "AKIAIOSFODNN7EXAMPLE"
    assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraRole"
    secret_key      = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  name          = "My AWS Credentials"
  provider_type = "aws"
  workspace_id  = 123
}