resource "seqera_orgs" "my_orgs" {
  logo_id = "...my_logo_id..."
  org_id  = 9
  organization = {
    description = "...my_description..."
    full_name   = "...my_full_name..."
    location    = "...my_location..."
    name        = "...my_name..."
    website     = "...my_website..."
  }
}