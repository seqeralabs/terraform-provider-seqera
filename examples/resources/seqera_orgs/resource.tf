# Seqera Organizations Resource Examples
#
# Organizations are the top-level structure in Seqera Platform that contain
# workspaces, members, and teams. Organizations provide multi-tenancy, resource
# isolation, and access control.

# Example 1: Basic organization
# Minimal configuration with required fields only

resource "seqera_orgs" "basic" {
  name      = "my-org"
  full_name = "My Organization"

  lifecycle {
    prevent_destroy = true
  }
}

# Example 2: Organization with optional metadata
# Include description, location, and website information

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
