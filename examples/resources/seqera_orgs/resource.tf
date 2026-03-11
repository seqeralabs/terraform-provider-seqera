resource "seqera_orgs" "basic" {
  name      = "my-org"
  full_name = "My Organization"

  lifecycle {
    prevent_destroy = true
  }
}
