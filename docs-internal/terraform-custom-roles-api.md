# Terraform Custom Roles — API Reference & Resource Design

This document covers the Seqera Platform APIs required to manage custom roles and role assignments via Terraform, along with a proposed Terraform resource design.

---

## Table of Contents

- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
  - [Custom Role CRUD](#1-custom-role-crud-roles)
  - [Role Assignment](#2-role-assignment-participants)
  - [Supporting / Read-Only](#3-supporting--read-only-endpoints)
- [Available Permissions (Grants)](#available-permissions-grants)
- [Predefined Workspace Roles](#predefined-workspace-roles)
- [Terraform Resource Design](#terraform-resource-design)
  - [seqera_custom_role](#seqera_custom_role)
  - [seqera_workspace_participant](#seqera_workspace_participant)
  - [Data Sources](#data-sources)
- [Example Usage](#example-usage)
- [Implementation Notes](#implementation-notes)

---

## Authentication

All endpoints require a Bearer token:

```
Authorization: Bearer <access-token>
```

OpenAPI security scheme: `BearerAuth`

---

## API Endpoints

### 1. Custom Role CRUD (`/roles`)

Base path: `/api/v1/roles`

#### List Available Permissions

Returns all grants that can be assigned to custom roles.

```
GET /roles/permissions?orgId={orgId}
```

**Query Parameters:**

| Param   | Type            | Required | Description               |
| ------- | --------------- | -------- | ------------------------- |
| `orgId` | integer (int64) | Yes      | Organization ID           |
| `name`  | string          | No       | Filter by permission name |

**Response** `200`:

```json
{
  "permissions": [
    { "name": "action:delete", "category": "Pipelines" },
    { "name": "action:execute", "category": "Pipelines" },
    { "name": "action:read", "category": "Pipelines" },
    { "name": "action:write", "category": "Pipelines" },
    { "name": "compute_environment:read", "category": "Compute" },
    { "name": "compute_environment:write", "category": "Compute" },
    { "name": "credentials:read", "category": "Compute" },
    ...
  ]
}
```

---

#### List Roles

Returns all roles (predefined + custom) in the organization.

```
GET /roles?orgId={orgId}
```

**Query Parameters:**

| Param    | Type            | Required | Description                       |
| -------- | --------------- | -------- | --------------------------------- |
| `orgId`  | integer (int64) | Yes      | Organization ID                   |
| `max`    | integer         | No       | Page size (default: 50, max: 100) |
| `offset` | integer         | No       | Pagination offset (default: 0)    |
| `name`   | string          | No       | Filter by role name               |
| `type`   | string          | No       | Filter: `predefined` or `custom`  |

**Response** `200`:

```json
{
  "roles": [
    {
      "name": "owner",
      "description": "Full permissions on organization resources...",
      "isPredefined": true
    },
    {
      "name": "admin",
      "description": "Full permissions on all workspace resources...",
      "isPredefined": true
    },
    {
      "name": "maintain",
      "description": "Full permissions on all workspace resources.",
      "isPredefined": true
    },
    {
      "name": "launch",
      "description": "Ability to launch pipelines...",
      "isPredefined": true
    },
    {
      "name": "connect",
      "description": "Ability to connect to running Studio sessions...",
      "isPredefined": true
    },
    {
      "name": "view",
      "description": "Ability to list, search, and view...",
      "isPredefined": true
    },
    {
      "name": "Pipeline Manager",
      "description": "Can manage pipelines and launch workflows",
      "isPredefined": false
    }
  ],
  "totalSize": 7
}
```

---

#### Describe Role

Returns a single role with its full permission set.

```
GET /roles/{roleName}?orgId={orgId}
```

**Path Parameters:**

| Param      | Type   | Description                           |
| ---------- | ------ | ------------------------------------- |
| `roleName` | string | The role name (URL-encoded if spaces) |

**Query Parameters:**

| Param   | Type            | Required | Description     |
| ------- | --------------- | -------- | --------------- |
| `orgId` | integer (int64) | Yes      | Organization ID |

**Response** `200`:

```json
{
  "role": {
    "name": "Pipeline Manager",
    "description": "Can manage pipelines and launch workflows",
    "isPredefined": false,
    "permissions": [
      "pipeline:read",
      "pipeline:write",
      "workflow:execute",
      "compute_environment:read"
    ]
  }
}
```

---

#### Create Role

```
POST /roles?orgId={orgId}
```

**Request Body:**

```json
{
  "name": "Pipeline Manager",
  "description": "Can manage pipelines and launch workflows",
  "permissions": [
    "pipeline:read",
    "pipeline:write",
    "workflow:execute",
    "compute_environment:read"
  ]
}
```

| Field         | Type     | Required | Constraints                                                |
| ------------- | -------- | -------- | ---------------------------------------------------------- |
| `name`        | string   | Yes      | Max 40 chars, must not conflict with predefined role names |
| `description` | string   | Yes      | Max 1000 chars, cannot be blank                            |
| `permissions` | string[] | Yes      | At least one; must be valid canonical grant names          |

**Response** `201`:

```json
{
  "name": "Pipeline Manager",
  "description": "Can manage pipelines and launch workflows",
  "permissions": [
    "pipeline:read",
    "pipeline:write",
    "workflow:execute",
    "compute_environment:read"
  ]
}
```

**Errors:**

- `400` — Invalid name, empty permissions, invalid grant names
- `403` — Insufficient permissions or feature not enabled
- `409` — Role name already exists in the organization

---

#### Update Role

```
PUT /roles/{roleName}?orgId={orgId}
```

**Request Body:** Same schema as Create.

**Response** `204` No Content.

**Errors:**

- `400` — Validation errors, or attempting to update a predefined role
- `403` — Insufficient permissions
- `404` — Role not found
- `409` — New name conflicts with existing role

---

#### Delete Role

```
DELETE /roles/{roleName}?orgId={orgId}
```

**Response** `204` No Content.

**Errors:**

- `400` — Role is still assigned to participants (must unassign first)
- `403` — Insufficient permissions
- `404` — Role not found

---

#### Validate Role Name

Check if a name is available before creating.

```
GET /roles/validate?name={roleName}&orgId={orgId}
```

**Response:**

- `204` — Name is valid and available
- `400` — Name is invalid (too long, bad characters, conflicts with predefined)
- `409` — Name already taken

---

### 2. Role Assignment (Participants)

These endpoints are **already in the OpenAPI spec**.

Base path: `/api/v1/orgs/{orgId}/workspaces/{workspaceId}/participants`

#### List Workspace Participants

```
GET /orgs/{orgId}/workspaces/{workspaceId}/participants
```

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `max` | integer | No | Page size |
| `offset` | integer | No | Pagination offset |
| `search` | string | No | Filter by name/email |

**Response** `200`:

```json
{
  "participants": [
    {
      "participantId": 101,
      "memberId": 42,
      "userName": "jdoe",
      "firstName": "Jane",
      "lastName": "Doe",
      "email": "jane@example.com",
      "orgRole": "member",
      "teamId": null,
      "teamName": null,
      "wspRole": "Pipeline Manager",
      "type": "MEMBER",
      "userAvatarUrl": "..."
    },
    {
      "participantId": 102,
      "memberId": null,
      "teamId": 7,
      "teamName": "Bioinformatics",
      "wspRole": "launch",
      "type": "TEAM",
      "teamAvatarUrl": "..."
    }
  ],
  "totalSize": 2
}
```

Note: `wspRole` is a string — it contains either a predefined role name (`owner`, `admin`, `maintain`, `launch`, `connect`, `view`) or a custom role name.

---

#### Add Participant to Workspace

```
PUT /orgs/{orgId}/workspaces/{workspaceId}/participants/add
```

**Request Body** (one of `memberId`, `teamId`, or `userNameOrEmail`):

```json
{ "memberId": 42 }
```

```json
{ "teamId": 7 }
```

```json
{ "userNameOrEmail": "jane@example.com" }
```

**Response** `200`:

```json
{
  "participant": {
    "participantId": 101,
    "memberId": 42,
    "wspRole": "launch",
    "type": "MEMBER",
    ...
  }
}
```

Note: Newly added participants get a default role. Use Update Participant Role to assign the desired role afterward.

---

#### Update Participant Role

```
PUT /orgs/{orgId}/workspaces/{workspaceId}/participants/{participantId}/role
```

**Request Body:**

```json
{
  "role": "Pipeline Manager"
}
```

The `role` field accepts either a predefined role name or a custom role name.

**Response** `204` No Content.

---

#### Delete Participant from Workspace

```
DELETE /orgs/{orgId}/workspaces/{workspaceId}/participants/{participantId}
```

**Response** `204` No Content.

---

### 3. Supporting / Read-Only Endpoints

These are useful for data sources in Terraform.

#### List Organization Members

```
GET /orgs/{orgId}/members
```

Returns members with their `memberId` (needed for adding participants).

#### List Organization Teams

```
GET /orgs/{orgId}/teams
```

Returns teams with their `teamId`.

#### List Workspaces

```
GET /orgs/{orgId}/workspaces
```

Returns workspaces with their `workspaceId`.

---

## Available Permissions (Grants)

Permissions use canonical names in `noun:verb` format. The full list is returned by `GET /roles/permissions`, but here is the reference grouped by category:

### Pipelines

| Permission                | Description                    |
| ------------------------- | ------------------------------ |
| `action:read`             | View pipeline actions          |
| `action:write`            | Create/update pipeline actions |
| `action:execute`          | Trigger pipeline actions       |
| `action:delete`           | Delete pipeline actions        |
| `action_label:write`      | Manage action labels           |
| `pipeline:read`           | View pipelines                 |
| `pipeline:write`          | Create/update pipelines        |
| `pipeline:delete`         | Delete pipelines               |
| `pipeline_label:write`    | Manage pipeline labels         |
| `pipeline_secrets:read`   | View pipeline secrets          |
| `pipeline_secrets:write`  | Create/update pipeline secrets |
| `pipeline_secrets:delete` | Delete pipeline secrets        |
| `launch:read`             | View launch configurations     |
| `workflow:read`           | View workflow runs             |
| `workflow:write`          | Update workflow metadata       |
| `workflow:execute`        | Launch workflows               |
| `workflow:delete`         | Delete workflow runs           |
| `workflow_label:write`    | Manage workflow labels         |
| `workflow_quick:execute`  | Quick-launch workflows         |
| `workflow_star:read`      | View starred workflows         |
| `workflow_star:write`     | Star workflows                 |
| `workflow_star:delete`    | Unstar workflows               |

### Compute

| Permission                            | Description                         |
| ------------------------------------- | ----------------------------------- |
| `compute_environment:read`            | View compute environments           |
| `compute_environment:write`           | Create/update compute environments  |
| `compute_environment:delete`          | Delete compute environments         |
| `credentials:read`                    | View credentials                    |
| `credentials:write`                   | Create/update credentials           |
| `credentials:delete`                  | Delete credentials                  |
| `credentials_encrypted:read`          | View encrypted credential details   |
| `container:read`                      | View container details              |
| `platform:read`                       | View compute platform info          |
| `managed_identity:read`               | View managed identities             |
| `managed_identity:write`              | Create/update managed identities    |
| `managed_identity:delete`             | Delete managed identities           |
| `managed_identity_credentials:read`   | View managed identity credentials   |
| `managed_identity_credentials:write`  | Update managed identity credentials |
| `managed_identity_credentials:delete` | Delete managed identity credentials |
| `managed_identity_credentials:admin`  | Admin managed identity credentials  |

### Data

| Permission            | Description                        |
| --------------------- | ---------------------------------- |
| `data_link:read`      | View data links                    |
| `data_link:write`     | Create/update data links           |
| `data_link:delete`    | Delete data links                  |
| `data_link:admin`     | Admin data links (manage metadata) |
| `dataset:read`        | View datasets                      |
| `dataset:write`       | Create/update datasets             |
| `dataset:delete`      | Delete datasets                    |
| `dataset:admin`       | Admin datasets                     |
| `dataset_label:write` | Manage dataset labels              |

### Studios

| Permission               | Description                |
| ------------------------ | -------------------------- |
| `studio:read`            | View Studios               |
| `studio:write`           | Create/update Studios      |
| `studio:execute`         | Start/stop Studios         |
| `studio:delete`          | Delete Studios             |
| `studio:admin`           | Admin Studios              |
| `studio_session:read`    | View Studio sessions       |
| `studio_session:execute` | Connect to Studio sessions |
| `studio_label:write`     | Manage Studio labels       |

### Settings

| Permission               | Description                      |
| ------------------------ | -------------------------------- |
| `label:read`             | View labels                      |
| `label:write`            | Create/update labels             |
| `label:delete`           | Delete labels                    |
| `workspace:read`         | View workspace details           |
| `workspace:write`        | Update workspace settings        |
| `workspace:delete`       | Delete workspace                 |
| `workspace:admin`        | Admin workspace settings         |
| `workspace_self:delete`  | Leave workspace                  |
| `workspace_studio:read`  | View workspace Studio settings   |
| `workspace_studio:write` | Update workspace Studio settings |

### Organizations

| Permission                 | Description                  |
| -------------------------- | ---------------------------- |
| `organization:read`        | View organization details    |
| `organization:write`       | Update organization settings |
| `organization:delete`      | Delete organization          |
| `organization_self:delete` | Leave organization           |
| `role:read`                | View roles                   |
| `role:write`               | Create/update custom roles   |
| `role:delete`              | Delete custom roles          |

### Cloud

| Permission              | Description                  |
| ----------------------- | ---------------------------- |
| `credits:read`          | View credits                 |
| `credits:admin`         | Manage credits               |
| `quota:read`            | View quotas                  |
| `quota:write`           | Manage quotas                |
| `eval_workspace:delete` | Delete evaluation workspaces |

---

## Predefined Workspace Roles

These are built-in and cannot be modified or deleted:

| Role       | Description                                                                        |
| ---------- | ---------------------------------------------------------------------------------- |
| `owner`    | Full permissions on organization resources, roles, and configuration               |
| `admin`    | Full permissions on all workspace resources, and ability to manage role assignment |
| `maintain` | Full permissions on all workspace resources                                        |
| `launch`   | Launch pipelines, modify parameters, connect to Studios                            |
| `connect`  | Connect to running Studio sessions and view workspace resources                    |
| `view`     | Read-only: list, search, and view status and configuration                         |

---

## Terraform Resource Design

### `seqera_custom_role`

Manages a custom role within an organization.

#### Schema

```hcl
resource "seqera_custom_role" "pipeline_manager" {
  organization_id = 12345
  name            = "Pipeline Manager"
  description     = "Can manage pipelines and launch workflows"

  permissions = [
    "pipeline:read",
    "pipeline:write",
    "pipeline:delete",
    "workflow:execute",
    "workflow:read",
    "compute_environment:read",
    "credentials:read",
  ]
}
```

#### Attributes

| Attribute         | Type        | Required | Description                              |
| ----------------- | ----------- | -------- | ---------------------------------------- |
| `organization_id` | number      | Yes      | Organization ID                          |
| `name`            | string      | Yes      | Role name (max 40 chars, unique per org) |
| `description`     | string      | Yes      | Role description (max 1000 chars)        |
| `permissions`     | set(string) | Yes      | Set of permission canonical names        |

#### Read-Only Attributes

| Attribute       | Type   | Description                          |
| --------------- | ------ | ------------------------------------ |
| `id`            | string | Composite ID: `{orgId}/{roleName}`   |
| `is_predefined` | bool   | Always `false` for managed resources |

#### API Mapping

| Terraform Op | API Call                                 |
| ------------ | ---------------------------------------- |
| Create       | `POST /roles?orgId={orgId}`              |
| Read         | `GET /roles/{roleName}?orgId={orgId}`    |
| Update       | `PUT /roles/{roleName}?orgId={orgId}`    |
| Delete       | `DELETE /roles/{roleName}?orgId={orgId}` |

#### Import

```bash
terraform import seqera_custom_role.pipeline_manager "12345/Pipeline Manager"
```

---

### `seqera_workspace_participant`

Manages a participant (user or team) in a workspace, including their role assignment.

#### Schema — Member Participant

```hcl
resource "seqera_workspace_participant" "jane_pipelines" {
  organization_id = 12345
  workspace_id    = 67890
  member_id       = 42
  role            = seqera_custom_role.pipeline_manager.name
}
```

#### Schema — Team Participant

```hcl
resource "seqera_workspace_participant" "bioinfo_team" {
  organization_id = 12345
  workspace_id    = 67890
  team_id         = 7
  role            = "launch"  # predefined role
}
```

#### Schema — Collaborator (external user by email)

```hcl
resource "seqera_workspace_participant" "external_viewer" {
  organization_id   = 12345
  workspace_id      = 67890
  user_name_or_email = "external@partner.com"
  role               = "view"
}
```

#### Attributes

| Attribute            | Type   | Required     | Description                         |
| -------------------- | ------ | ------------ | ----------------------------------- |
| `organization_id`    | number | Yes          | Organization ID                     |
| `workspace_id`       | number | Yes          | Workspace ID                        |
| `member_id`          | number | One of three | Org member ID                       |
| `team_id`            | number | One of three | Team ID                             |
| `user_name_or_email` | string | One of three | Username or email for collaborators |
| `role`               | string | Yes          | Predefined or custom role name      |

Exactly one of `member_id`, `team_id`, or `user_name_or_email` must be set.

#### Read-Only Attributes

| Attribute        | Type   | Description                                        |
| ---------------- | ------ | -------------------------------------------------- |
| `id`             | string | Composite: `{orgId}/{workspaceId}/{participantId}` |
| `participant_id` | number | Server-assigned participant ID                     |
| `type`           | string | `MEMBER`, `TEAM`, or `COLLABORATOR`                |

#### API Mapping

| Terraform Op         | API Call                                                                                            |
| -------------------- | --------------------------------------------------------------------------------------------------- |
| Create               | `PUT /orgs/{orgId}/workspaces/{workspaceId}/participants/add` then `PUT .../participants/{id}/role` |
| Read                 | `GET /orgs/{orgId}/workspaces/{workspaceId}/participants` (filter by participant ID)                |
| Update (role change) | `PUT .../participants/{participantId}/role`                                                         |
| Delete               | `DELETE .../participants/{participantId}`                                                           |

Note: Create is a two-step operation — the add endpoint creates the participant with a default role, then the role update endpoint sets the desired role.

#### Import

```bash
terraform import seqera_workspace_participant.jane_pipelines "12345/67890/101"
```

---

### Data Sources

#### `seqera_custom_roles` (list)

```hcl
data "seqera_custom_roles" "all" {
  organization_id = 12345
}

# data.seqera_custom_roles.all.roles[*].name
# data.seqera_custom_roles.all.roles[*].permissions
```

API: `GET /roles?orgId={orgId}&type=custom`

#### `seqera_custom_role` (single)

```hcl
data "seqera_custom_role" "pipeline_mgr" {
  organization_id = 12345
  name            = "Pipeline Manager"
}

# data.seqera_custom_role.pipeline_mgr.permissions
```

API: `GET /roles/{roleName}?orgId={orgId}`

#### `seqera_role_permissions` (available grants)

```hcl
data "seqera_role_permissions" "available" {
  organization_id = 12345
}

# data.seqera_role_permissions.available.permissions[*].name
# data.seqera_role_permissions.available.permissions[*].category
```

API: `GET /roles/permissions?orgId={orgId}`

---

## Example Usage

### Full example: custom roles with team and user assignments

```hcl
terraform {
  required_providers {
    seqera = {
      source = "seqera/seqera"
    }
  }
}

provider "seqera" {
  api_url = "https://api.cloud.seqera.io"
  token   = var.seqera_token
}

# --- Custom Roles ---

resource "seqera_custom_role" "pipeline_manager" {
  organization_id = var.org_id
  name            = "Pipeline Manager"
  description     = "Manages pipelines, launches workflows, views compute and credentials"

  permissions = [
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

resource "seqera_custom_role" "data_steward" {
  organization_id = var.org_id
  name            = "Data Steward"
  description     = "Manages datasets, data links, and can view pipelines"

  permissions = [
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

resource "seqera_custom_role" "studio_user" {
  organization_id = var.org_id
  name            = "Studio User"
  description     = "Can create and use Studios, view pipelines"

  permissions = [
    "studio:read",
    "studio:write",
    "studio:execute",
    "studio_session:read",
    "studio_session:execute",
    "pipeline:read",
    "workflow:read",
    "compute_environment:read",
  ]
}

# --- Assign users to workspaces with custom roles ---

resource "seqera_workspace_participant" "jane_pipeline_mgr" {
  organization_id = var.org_id
  workspace_id    = var.production_workspace_id
  member_id       = var.jane_member_id
  role            = seqera_custom_role.pipeline_manager.name
}

resource "seqera_workspace_participant" "bob_data_steward" {
  organization_id = var.org_id
  workspace_id    = var.production_workspace_id
  member_id       = var.bob_member_id
  role            = seqera_custom_role.data_steward.name
}

# --- Assign a team with a predefined role ---

resource "seqera_workspace_participant" "bioinfo_team" {
  organization_id = var.org_id
  workspace_id    = var.production_workspace_id
  team_id         = var.bioinfo_team_id
  role            = "maintain"
}

# --- Assign a team with a custom role ---

resource "seqera_workspace_participant" "ml_team_studios" {
  organization_id = var.org_id
  workspace_id    = var.ml_workspace_id
  team_id         = var.ml_team_id
  role            = seqera_custom_role.studio_user.name
}
```

---

## Implementation Notes

### Role identification

Roles are identified by **name** (not numeric ID) in all API interactions. The name is unique per organization (case-insensitive via slug). This means:

- Terraform resource IDs should be composite: `{orgId}/{roleName}`
- Renaming a role via `PUT` changes its identity — the provider should handle this with `ForceNew` or by using the old name in the URL and the new name in the body

### Two-step participant creation

Adding a participant and setting their role are separate API calls. The provider's Create function must:

1. `PUT /orgs/{orgId}/workspaces/{workspaceId}/participants/add` — creates with default role
2. `PUT .../participants/{participantId}/role` — sets the desired role

### Participant lookup

There is no `GET /participants/{id}` endpoint. To read a specific participant, use `GET /participants?search=...` and filter by `participantId` from the list. The provider should paginate through results if necessary.

### Role deletion prerequisite

A custom role cannot be deleted while assigned to any participants. The provider should document this — users must remove or reassign all participants using the role before destroying it. Consider using `depends_on` or lifecycle ordering in Terraform configurations.

### Quota limit

Organizations have a configurable limit on custom roles (default: 20, set via `tower.limits.max-custom-roles`). The provider should surface the `400` error clearly when this limit is reached.

### Feature gate

Custom roles require the `CUSTOM_ROLES` feature to be enabled. Self-hosted: always enabled. Cloud: depends on organization tier (not available on Basic). The `403` from `FeatureCustomRolesEnabledChecker` should be surfaced as a clear error message.
