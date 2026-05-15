---
page_title: "Fine-grained access control with custom roles"
subcategory: "Examples"
description: |-
  Manage Seqera Platform custom roles, look up teams by name, and assign roles to workspace participants — with plan-time validation against the live permission catalogue.
---

# Fine-grained access control with custom roles

Seqera Platform ships with six predefined workspace roles (`owner`, `admin`, `maintain`, `launch`, `connect`, `view`) that cover the common case but can be too coarse for organizations that want to separate, for example, pipeline editors from data stewards. **Custom roles** let an organization admin define a named permission set and assign it to workspace participants alongside the predefined roles.

For an overview of the feature and the full permission reference, see [Custom roles](https://docs.seqera.io/platform-cloud/orgs-and-teams/custom-roles) in the Seqera Platform docs. This guide is the Terraform-side companion to that page.

~> **Availability.** Custom roles require **Seqera Platform Cloud Pro** or **Seqera Platform Enterprise v25.3 or later**. On other tiers / older Enterprise versions the API returns HTTP 403 on create — see [Behaviours worth knowing](#behaviours-worth-knowing).

This guide covers:

- The four Terraform surfaces that ship for custom roles
- A worked example: a Pipeline Manager role assigned to a team
- The plan-time validation pattern that catches typos in permission names before they hit the API
- Behaviors worth knowing about: rename semantics and the tier requirement

## The four surfaces

| Surface | Type | Purpose |
|---|---|---|
| `seqera_custom_role` | resource | CRUD on a named permission set in an organization |
| `seqera_team` | data source | Look up a team by name → returns `team_id` |
| `seqera_custom_role` | data source | Look up a role (predefined or custom) → returns `permissions`, `is_predefined` |
| `seqera_permissions` | data source | Live catalogue of all assignable permissions — used for validation |

Custom roles attach to participants via the existing [`seqera_workspace_participant`](../resources/workspace_participant.md) resource. The `role` argument accepts either a predefined role name or the name of a custom role in the same organization.

## End-to-end example

```terraform
terraform {
  required_providers {
    seqera = { source = "seqeralabs/seqera" }
  }
}

data "seqera_organization" "main" {
  name = "my-organization"
}

# Live catalogue, used as a plan-time guard against unknown permission names.
data "seqera_permissions" "all" {
  org_id = data.seqera_organization.main.org_id
}

# Resolve the engineering team's team_id by name — avoids hardcoding numeric IDs.
data "seqera_team" "engineering" {
  org_id = data.seqera_organization.main.org_id
  name   = "engineering"
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
}

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
      error_message = <<-EOT
        One or more permissions on `pipeline_manager` is not in the live platform
        catalogue. Bad entries: ${jsonencode([
        for p in local.pipeline_manager_permissions : p if !contains(data.seqera_permissions.all.names, p)
])}
      EOT
    }
  }
}

# Assign the team to a workspace under the custom role.
resource "seqera_workspace_participant" "engineering_pipeline_mgr" {
  org_id       = data.seqera_organization.main.org_id
  workspace_id = seqera_workspace.production.id
  team_id      = data.seqera_team.engineering.team_id
  role         = seqera_custom_role.pipeline_manager.name
}
```

Apply produces:

- One custom role on the platform (`Pipeline Manager`, with 11 permissions, `isPredefined = false`).
- One participant record on the production workspace, with the role string set to `Pipeline Manager`.

Removing or renaming the role later requires reassigning participants — see [Behaviours worth knowing](#behaviours-worth-knowing) below.

## Plan-time permission validation

Permission names follow a `{resource}:{verb}` convention (`pipeline:read`, `workflow:execute`, `compute_environment:write`, and so on). A typo causes the API to return HTTP 400 — but only after apply has started talking to the server. The `seqera_permissions` data source returns the live catalogue, which lets you reject unknown permission names at plan time before any API call:

```terraform
data "seqera_permissions" "all" {
  org_id = data.seqera_organization.main.org_id
}

locals {
  permissions = [
    "pipeline:read",
    "pipeline:reed",  # invalid — caught at plan time
    "workflow:run",   # invalid — caught at plan time
  ]
}

resource "seqera_custom_role" "guarded" {
  # ...
  permissions = local.permissions

  lifecycle {
    precondition {
      condition = alltrue([
        for p in local.permissions : contains(data.seqera_permissions.all.names, p)
      ])
      error_message = "Bad permissions: ${jsonencode([
        for p in local.permissions : p if !contains(data.seqera_permissions.all.names, p)
      ])}"
    }
  }
}
```

A `terraform plan` with the invalid entries above fails with:

```
Error: Resource precondition failed

  on main.tf line 23, in resource "seqera_custom_role" "guarded":

Bad permissions: ["pipeline:reed","workflow:run"]
```

No state is touched and no API call is made.

### Why a `local` rather than `self.permissions`

Terraform's `precondition` block doesn't have access to the `self.*` value (it isn't part of the planned state yet). Extracting the permission list to a `local` and referencing the local both in the resource argument and the precondition condition is the workaround. `self.*` is only available inside `postcondition` and provisioner blocks.

## Filtering the catalogue

The `seqera_permissions` data source accepts an optional `category` filter:

```terraform
data "seqera_permissions" "pipelines_only" {
  org_id   = data.seqera_organization.main.org_id
  category = "Pipelines"
}

output "pipeline_perms_count" {
  value = length(data.seqera_permissions.pipelines_only.permissions)
}
```

Categories observed in the platform catalogue: `Compute`, `Data`, `Pipelines`, `Settings`, `Studios`. The filter is case-insensitive and applied client-side.

## Reading a role's permission set

The `seqera_custom_role` data source works for **both** custom and predefined roles. Useful for introspecting what `maintain` actually grants:

```terraform
data "seqera_custom_role" "maintain" {
  org_id = data.seqera_organization.main.org_id
  name   = "maintain"
}

output "maintain_permissions" {
  value = data.seqera_custom_role.maintain.permissions
}

output "maintain_is_predefined" {
  value = data.seqera_custom_role.maintain.is_predefined  # true
}
```

This is the cleanest way to derive a custom role from a predefined one — copy the `permissions` list, drop or add the few you need, and use it as the basis for `seqera_custom_role.permissions`.

## Behaviours worth knowing

### Rename is force-new

Changing `seqera_custom_role.name` destroys the role and creates a new one. The destroy step removes the role from any workspace participants still assigned to it, so those participants need to be reassigned via the `seqera_workspace_participant.role` argument before the apply, or you'll see participants flip to the workspace's default role.

The platform API _does_ support in-place rename (`PUT /roles/{old}` with `body.name = new` preserves assignments), but the current provider uses force-new for predictability. Future releases may change this behaviour.

### Tier requirement

Custom roles require Seqera Platform Cloud Pro or Seqera Platform Enterprise v25.3 or later. Applying `seqera_custom_role` against an organization without the entitlement returns HTTP 403 on create. If you're unsure whether the feature is enabled for your org, the **Access control** section under the org switcher in the Platform UI will show an **Add role** button on tiers that have it.

To find out more, contact your Seqera representative or [reach out via the Seqera support portal](https://support.seqera.io/support/home).

### Permission names match `{resource}:{verb}`

The verb is one of: `read`, `write`, `delete`, `execute`, `admin`. The `admin` verb is reserved for **cross-user** privileged operations (e.g. `studio:admin` = manage another user's studio session, `workspace:admin` = transfer the Owner role). Most custom roles should not need `:admin` grants.

`*_label:write` permissions (e.g. `pipeline_label:write`, `dataset_label:write`) are split out from the parent `:write` so a role can grant labelling rights without granting full edit. Useful for ops or data-steward archetypes.

For the canonical list of every permission and the API endpoints it gates, see the [Permissions table](https://docs.seqera.io/platform-cloud/orgs-and-teams/custom-roles#permissions) in the Seqera docs.

### Predefined-role parity

Predefined roles (`owner`, `admin`, `maintain`, `launch`, `connect`, `view`) cannot be modified. They show up in `seqera_custom_role` data-source lookups with `is_predefined = true`. Custom roles created with the same name as a predefined role are rejected by the API.

## Archetype roles

Three common starting points, all derived from the permission catalogue. Use them as templates and trim to your needs.

```terraform
# Pipeline Manager — can manage pipelines and launch runs.
locals {
  pipeline_manager_permissions = [
    "pipeline:read", "pipeline:write", "pipeline:delete", "pipeline_label:write",
    "workflow:read", "workflow:execute", "workflow:delete", "workflow_label:write",
    "compute_environment:read", "credentials:read", "label:read",
  ]
}

# Data Steward — manage datasets and data-links, view pipelines.
locals {
  data_steward_permissions = [
    "dataset:read", "dataset:write", "dataset:delete", "dataset:admin", "dataset_label:write",
    "data_link:read", "data_link:write", "data_link:delete", "data_link:admin",
    "pipeline:read", "workflow:read", "label:read",
  ]
}

# Studio User — create and use Studios, view pipelines.
locals {
  studio_user_permissions = [
    "studio:read", "studio:write", "studio:execute",
    "studio_session:read", "studio_session:execute",
    "pipeline:read", "workflow:read", "compute_environment:read",
  ]
}
```
