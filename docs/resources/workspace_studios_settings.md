---
page_title: "seqera_workspace_studios_settings Resource - terraform-provider-seqera"
subcategory: "Studios"
description: |-
  Manage the Studios settings of a Seqera workspace.
  These settings apply to Studios created in the workspace: the maximum session lifespan, whether new Studios are private by default, and where Wave pushes the images it builds for Studios.
  ~> Note: Platform has no way to delete these settings. Destroying this resource only removes it from Terraform state and leaves the workspace's current settings in place.
  ~> Note: When lifespan_hours is not 0 (unlimited), Platform only accepts Studios in the workspace whose configuration.lifespan_hours equals this value.
---

# seqera_workspace_studios_settings (Resource)

Manage the Studios settings of a Seqera workspace.

These settings apply to Studios created in the workspace: the maximum session lifespan, whether new Studios are private by default, and where Wave pushes the images it builds for Studios.

~> **Note:** Platform has no way to delete these settings. Destroying this resource only removes it from Terraform state and leaves the workspace's current settings in place.

~> **Note:** When `lifespan_hours` is not `0` (unlimited), Platform only accepts Studios in the workspace whose `configuration.lifespan_hours` equals this value.

## Example Usage

```terraform
# Studios settings for a workspace. Platform stores a single set of settings per
# workspace, so only one of these resources should manage a given workspace.
#
# The Platform API replaces all settings on every change, so lifespan_hours and
# private_studio_by_default are always required. Unset optional fields fall back
# to the platform defaults.
#
# Destroying this resource only removes it from Terraform state: Platform has no
# way to delete the settings, so the workspace keeps its current values.
resource "seqera_workspace_studios_settings" "main" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  # Maximum session lifespan in hours; 0 means unlimited. When this is not 0,
  # Platform only accepts Studios in the workspace whose
  # configuration.lifespan_hours equals this value.
  lifespan_hours = 8

  # New Studios are private to their creator unless is_private says otherwise.
  private_studio_by_default = true

  # Where Wave pushes images it builds for Studios (for example from a
  # remote_config or conda environment), and how those images are named.
  container_repository = "cr.seqera.io/my-org/studios"
  name_strategy        = "tagPrefix" # none | tagPrefix | imageSuffix
}
```

### Fixed Lifespan

```terraform
# Enforce one lifespan across the workspace. Because lifespan_hours is not 0,
# Platform rejects Studios whose configuration.lifespan_hours differs, so the
# Studio below reads the value from the settings resource instead of repeating it.
resource "seqera_workspace_studios_settings" "fixed" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  lifespan_hours            = 8
  private_studio_by_default = true
}

resource "seqera_studios" "matches_workspace" {
  name                 = "studio-matching-workspace"
  compute_env_id       = seqera_compute_env.main.id
  workspace_id         = seqera_workspace.main.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"

  configuration = {
    lifespan_hours = seqera_workspace_studios_settings.fixed.lifespan_hours
  }
}
```

### Minimal

```terraform
# The smallest valid configuration: only the two settings the Platform API
# always needs. container_repository and name_strategy are left unset, so Wave
# uses the platform defaults for where it pushes Studio images and how it names
# them.
resource "seqera_workspace_studios_settings" "minimal" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  lifespan_hours            = 8
  private_studio_by_default = false
}
```

### Unlimited Lifespan

```terraform
# lifespan_hours = 0 removes the limit on Studio session length. This is also
# the only setting that lets each Studio choose its own
# `configuration.lifespan_hours`.
#
# With any other value, Platform only accepts Studios in the workspace whose
# configuration.lifespan_hours equals the workspace value (see the next example).
resource "seqera_workspace_studios_settings" "unlimited" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  lifespan_hours            = 0
  private_studio_by_default = true

  container_repository = "cr.seqera.io/my-org/studios"
  name_strategy        = "imageSuffix"
}

resource "seqera_studios" "custom_lifespan" {
  name                 = "studio-custom-lifespan"
  compute_env_id       = seqera_compute_env.main.id
  workspace_id         = seqera_workspace.main.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"

  configuration = {
    lifespan_hours = 24
  }

  # The settings must exist first so the limit is in place when the Studio is
  # created.
  depends_on = [seqera_workspace_studios_settings.unlimited]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `lifespan_hours` (Number) Maximum number of hours a Studio session may run before it is automatically stopped. Use 0 for unlimited lifespan.
- `org_id` (Number) Organization numeric identifier. Requires replacement if changed.
- `private_studio_by_default` (Boolean) When true, newly created Studio sessions are private by default and only accessible to their creator.
- `workspace_id` (Number) Workspace numeric identifier. Requires replacement if changed.

### Optional

- `container_repository` (String) Container repository where Wave pushes images built for Studios in this workspace (e.g. 'cr.seqera.io/my-org/studios'). When null, the platform default is used.
- `name_strategy` (String) Wave image naming strategy applied to container images built for Studios in this workspace. When null, the platform default strategy is used. must be one of ["none", "tagPrefix", "imageSuffix"]

## Import

Import is supported using the following syntax:

In Terraform v1.5.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `id` attribute, for example:

```terraform
import {
  to = seqera_workspace_studios_settings.my_seqera_workspace_studios_settings
  id = jsonencode({
    org_id       = 0
    workspace_id = 0
  })
}
```

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import seqera_workspace_studios_settings.my_seqera_workspace_studios_settings '{"org_id": 0, "workspace_id": 0}'
```
