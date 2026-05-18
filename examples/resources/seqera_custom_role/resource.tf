# Live catalogue of permissions — used below to validate each role's
# permission list against the platform's canonical set at plan time.
data "seqera_permissions" "all" {
  org_id = data.seqera_organization.main.org_id
}

locals {
  pipeline_manager_permissions = [
    "pipeline:read",
    "pipeline:write",
    "pipeline:delete",
    "pipeline_label:write",
    "workflow:read",
    "workflow:execute",
    "workflow:delete",
    "workflow_label:write",
    "compute_environment:read",
    "credentials:read",
    "label:read",
  ]

  data_steward_permissions = [
    "dataset:read",
    "dataset:write",
    "dataset:delete",
    "dataset:admin",
    "dataset_label:write",
    "data_link:read",
    "data_link:write",
    "data_link:delete",
    "data_link:admin",
    "pipeline:read",
    "workflow:read",
    "label:read",
  ]
}

# Custom role: Pipeline Manager — manage pipelines and launch workflows.
resource "seqera_custom_role" "pipeline_manager" {
  org_id      = data.seqera_organization.main.org_id
  name        = "Pipeline Manager"
  description = "Manage pipelines, launch workflows, view compute and credentials"
  permissions = local.pipeline_manager_permissions

  lifecycle {
    precondition {
      condition = alltrue([
        for p in local.pipeline_manager_permissions : contains(data.seqera_permissions.all.names, p)
      ])
      error_message = "One or more permissions on `pipeline_manager` is not in the live platform catalogue. Check `data.seqera_permissions.all.names`."
    }
  }
}

# Custom role: Data Steward — manage datasets and data-links.
resource "seqera_custom_role" "data_steward" {
  org_id      = data.seqera_organization.main.org_id
  name        = "Data Steward"
  description = "Manage datasets and data-links; read-only view of pipelines and workflows"
  permissions = local.data_steward_permissions

  lifecycle {
    precondition {
      condition = alltrue([
        for p in local.data_steward_permissions : contains(data.seqera_permissions.all.names, p)
      ])
      error_message = "One or more permissions on `data_steward` is not in the live platform catalogue. Check `data.seqera_permissions.all.names`."
    }
  }
}

# Assign the custom role to a team via the existing workspace_participant
# resource. The role name on the participant resolves to either a
# predefined or custom role.
resource "seqera_workspace_participant" "engineering_pipeline_mgr" {
  org_id       = data.seqera_organization.main.org_id
  workspace_id = seqera_workspace.main.id
  team_id      = data.seqera_team.engineering.team_id
  role         = seqera_custom_role.pipeline_manager.name
}
