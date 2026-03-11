resource "seqera_orgs" "research" {
  name        = "research-lab"
  full_name   = "Research Laboratory"
  description = "Organization for computational research"
  location    = "San Francisco, CA"
  website     = "https://www.research-lab.org"

  lifecycle {
    prevent_destroy = true
  }
}
