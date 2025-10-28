resource "seqera_codecommit_credential" "my_codecommitcredential" {
  access_key   = "AKIAIOSFODNN7EXAMPLE"
  base_url     = "https://git-codecommit.us-east-1.amazonaws.com"
  name         = "...my_name..."
  secret_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  workspace_id = 9
}