data "seqera_managed_credentials" "my_managedcredentials" {
  managed_identity_id = 10
  max                 = 7
  offset              = 5
  org_id              = 10
  search              = "...my_search..."
  user_id             = 5
}