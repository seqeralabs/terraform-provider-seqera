# Look up an organization by name
data "seqera_orgs" "example" {
  name = "my-organization"
}

# Use the organization ID to look up workspaces
data "seqera_workspace" "example" {
  org_id = data.seqera_orgs.example.id
  name   = "my-workspace"
}

# Output the organization details
output "org_id" {
  value = data.seqera_orgs.example.id
}

output "org_full_name" {
  value = data.seqera_orgs.example.full_name
}

output "org_description" {
  value = data.seqera_orgs.example.description
}
