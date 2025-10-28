# Seqera Organizations Resource Examples
#
# Organizations are the top-level structure in Seqera Platform that contain
# workspaces, members, and teams. Organizations provide multi-tenancy, resource
# isolation, and access control.
#
# KEY CONCEPTS:
# - Organizations contain multiple workspaces
# - Each organization has members with roles (owner, member, collaborator)
# - Teams can be created within organizations for group access control
# - Workspaces within an organization can share resources

# Example 1: Basic organization
# Minimal configuration with required fields only

resource "seqera_orgs" "basic" {
  name      = "my-org"
  full_name = "My Organization"
}

# Example 2: Complete organization with all optional fields
# Full metadata including location, website, and description

resource "seqera_orgs" "complete" {
  name        = "biotech-research"
  full_name   = "Biotech Research Institute"
  description = "Organization for genomics and computational biology research"
  location    = "Boston, MA"
  website     = "https://www.biotechresearch.org"
}

# Example 3: Organization with workspaces
# Create an organization and associated workspaces

resource "seqera_orgs" "research" {
  name        = "research-lab"
  full_name   = "Research Laboratory"
  description = "Multi-project research organization"
  location    = "San Francisco, CA"
}

resource "seqera_workspace" "analysis" {
  name        = "analysis-workspace"
  org_id      = seqera_orgs.research.org_id
  full_name   = "${seqera_orgs.research.name}/analysis-workspace"
  description = "Workspace for data analysis workflows"
  visibility  = "PRIVATE"
}

resource "seqera_workspace" "production" {
  name        = "production-workspace"
  org_id      = seqera_orgs.research.org_id
  full_name   = "${seqera_orgs.research.name}/production-workspace"
  description = "Production workflows"
  visibility  = "PRIVATE"
}

# Example 4: Organization with teams
# Create an organization and teams for access control

resource "seqera_orgs" "enterprise" {
  name      = "enterprise-org"
  full_name = "Enterprise Organization"
}

resource "seqera_teams" "data_science" {
  name        = "data-science"
  description = "Data science team"
  org_id      = seqera_orgs.enterprise.org_id
}

resource "seqera_teams" "devops" {
  name        = "devops"
  description = "DevOps and infrastructure team"
  org_id      = seqera_orgs.enterprise.org_id
}

# Example 5: Multi-environment setup
# Organization with workspaces for different environments

resource "seqera_orgs" "multi_env" {
  name        = "multi-env-org"
  full_name   = "Multi-Environment Organization"
  description = "Organization with dev/staging/prod environments"
}

locals {
  environments = ["dev", "staging", "prod"]
}

resource "seqera_workspace" "environments" {
  for_each = toset(local.environments)

  name        = "${each.value}-workspace"
  org_id      = seqera_orgs.multi_env.org_id
  full_name   = "${seqera_orgs.multi_env.name}/${each.value}-workspace"
  description = "${title(each.value)} environment workspace"
  visibility  = each.value == "prod" ? "PRIVATE" : "SHARED"
}
