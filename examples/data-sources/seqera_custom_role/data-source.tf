# Look up a role (predefined or custom) by name.
data "seqera_organization" "main" {
  name = "my-organization"
}

# A custom role managed elsewhere — for example, by another Terraform
# config or created via the Seqera UI.
data "seqera_custom_role" "pipeline_manager" {
  org_id = data.seqera_organization.main.org_id
  name   = "Pipeline Manager"
}

# Predefined roles work too — useful for introspecting the permission
# set behind built-in roles like `maintain` or `launch`.
data "seqera_custom_role" "maintain" {
  org_id = data.seqera_organization.main.org_id
  name   = "maintain"
}

output "pipeline_manager_permissions" {
  value = data.seqera_custom_role.pipeline_manager.permissions
}

output "maintain_is_predefined" {
  value = data.seqera_custom_role.maintain.is_predefined
}
