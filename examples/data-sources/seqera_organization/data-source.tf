# Look up an organization by name
data "seqera_organization" "example" {
  name = "my-organization"
}

# Use the organization ID to look up workspaces
data "seqera_workspace" "example" {
  org_id = data.seqera_organization.example.org_id
  name   = "my-workspace"
}

# Output the organization details
output "org_id" {
  value = data.seqera_organization.example.org_id
}

output "org_full_name" {
  value = data.seqera_organization.example.full_name
}

output "org_description" {
  value = data.seqera_organization.example.description
}
