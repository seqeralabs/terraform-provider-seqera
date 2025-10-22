resource "seqera_managed_credentials" "my_managedcredentials" {
  checked                = false
  credential_provider    = "ssh"
  managed_credentials_id = 5
  managed_identity_id    = 10
  org_id                 = 10
  user_id                = 5
}