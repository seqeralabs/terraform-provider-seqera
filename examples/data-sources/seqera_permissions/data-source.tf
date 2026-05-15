# Catalogue of permissions available for custom roles.
data "seqera_organization" "main" {
  name = "my-organization"
}

# Full catalogue
data "seqera_permissions" "all" {
  org_id = data.seqera_organization.main.org_id
}

# Filtered to just one category
data "seqera_permissions" "pipelines" {
  org_id   = data.seqera_organization.main.org_id
  category = "Pipelines"
}

locals {
  pipeline_manager_permissions = [
    "pipeline:read",
    "pipeline:write",
    "workflow:read",
    "workflow:execute",
  ]
}

# Validate at plan time that every permission used by a custom role
# is in the live catalogue.
resource "seqera_custom_role" "pipeline_manager" {
  org_id      = data.seqera_organization.main.org_id
  name        = "Pipeline Manager"
  description = "Manage pipelines, launch workflows"
  permissions = local.pipeline_manager_permissions

  lifecycle {
    precondition {
      condition = alltrue([
        for p in local.pipeline_manager_permissions : contains(data.seqera_permissions.all.names, p)
      ])
      error_message = "One or more permissions on this role is not in the live platform catalogue."
    }
  }
}

output "all_permission_names" {
  value = data.seqera_permissions.all.names
}

output "categories" {
  value = data.seqera_permissions.all.categories
}

output "pipeline_only_count" {
  value = length(data.seqera_permissions.pipelines.permissions)
}
