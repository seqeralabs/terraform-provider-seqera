---
page_title: "seqera_studios Resource - terraform-provider-seqera"
subcategory: "Studios"
description: |-
  Studios is a unified platform where you can host a combination of
  container images and compute environments for interactive analysis using
  your preferred tools, like JupyterLab, an R-IDE, Visual Studio Code IDEs,
  or Xpra remote desktops. Each Studio session is an individual interactive
  environment that encapsulates the live environment for dynamic data analysis.
  Note:
  On Seqera Cloud, the free tier permits only one running Studio session at a time.
  To run simultaneous sessions, contact Seqera for a Seqera Cloud Pro license.
---

# seqera_studios (Resource)

Studios is a unified platform where you can host a combination of
container images and compute environments for interactive analysis using
your preferred tools, like JupyterLab, an R-IDE, Visual Studio Code IDEs,
or Xpra remote desktops. Each Studio session is an individual interactive
environment that encapsulates the live environment for dynamic data analysis.

Note:
On Seqera Cloud, the free tier permits only one running Studio session at a time.
To run simultaneous sessions, contact Seqera for a Seqera Cloud Pro license.

## Example Usage

```terraform
resource "seqera_studios" "basic_jupyter" {
  name                 = "my-jupyter-studio"
  compute_env_id       = seqera_compute_env.main.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  workspace_id         = seqera_workspace.main.id
  # Configuration is required - gpu defaults to 0
  configuration = {}
}
```

### Allowed Users

```terraform
# Grant another user access to a private Studio.
#
# The Studio owner is always the identity running Terraform: the studio is
# created under the account that owns the TOWER_ACCESS_TOKEN (or provider
# `bearer_auth`) credential. That owner is always allowed to connect and start
# the Studio and never needs to be listed in `allowed_user_ids` — if you include
# the owner's ID, the platform silently drops it from the list.
#
# `allowed_user_ids` takes Seqera *user IDs*, and is only valid on a private
# studio (is_private = true). Each allowed user must already be a member of the
# Studio's workspace, otherwise the platform rejects the request. Currently the
# allow list is capped at one additional user.

# Look up the user's numeric user ID from their email. `seqera_organization_member`
# exposes `user_id`, which is the value `allowed_user_ids` expects.
# (The `seqera_workspace_participant` data source only returns member/participant
# IDs, not the user ID, so it is not used here.)
data "seqera_organization_member" "collaborator" {
  org_id = seqera_workspace.main.org_id
  email  = "collaborator@example.com"
}

resource "seqera_studios" "shared_private" {
  name                 = "shared-private-studio"
  compute_env_id       = seqera_compute_env.main.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  workspace_id         = seqera_workspace.main.id
  configuration        = {}

  # Private studio: only the owner (the Terraform identity) and the users listed
  # below can connect to and start it.
  is_private       = true
  allowed_user_ids = [data.seqera_organization_member.collaborator.user_id]
}

# The resolved membership is available read-only, including the owner:
output "studio_allowed_users" {
  value = seqera_studios.shared_private.allowed_users
}
```

### Conda Heredoc

```terraform
resource "seqera_studios" "jupyter_with_conda_heredoc" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
  configuration = {
    # Use heredoc for simple YAML - just copy/paste your conda environment
    conda_environment = <<-EOT
      channels:
        - conda-forge
        - bioconda
      dependencies:
        - numpy>1.7,<2.3
        - scipy
        - tqdm=4.*
        - pip:
          - matplotlib==3.10.*
          - seaborn>=0.13
    EOT
    cpu               = 2
    memory            = 4096
    lifespan_hours    = 8
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  description          = "Jupyter studio with conda packages defined using heredoc"
  is_private           = true
  name                 = "jupyter-with-conda-heredoc"
  spot                 = true
  workspace_id         = seqera_workspace.main.id
}
```

### Conda Yamlencode Labels

```terraform
resource "seqera_studios" "jupyter_with_conda_labels" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
  configuration = {
    # Use yamlencode() for dynamic generation or when using Terraform variables
    conda_environment = yamlencode({
      channels = [
        "conda-forge",
        "bioconda"
      ]
      dependencies = [
        "numpy>1.7,<2.3",
        "scipy",
        "tqdm=4.*",
        {
          pip = [
            "matplotlib==3.10.*",
            "seaborn>=0.13"
          ]
        }
      ]
    })
    cpu            = 2
    memory         = 4096
    lifespan_hours = 8
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  description          = "Jupyter studio for data analysis and visualization"
  is_private           = true
  # Reference label IDs from seqera_labels resources
  label_ids = [
    seqera_labels.environment_prod.id,
    seqera_labels.team_datascience.id
  ]
  name         = "jupyter-with-conda-labels"
  spot         = true
  workspace_id = seqera_workspace.main.id
}
```

### Env Vars

```terraform
resource "seqera_studios" "studio_with_env_vars" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
  configuration = {
    cpu            = 2
    memory         = 8192
    lifespan_hours = 8
    # Studio-specific environment variables (keys must be alphanumeric + underscore, cannot start with number)
    environment = {
      MY_STUDIO_VAR = "testing"
      API_ENDPOINT  = "https://api.example.com"
      DEBUG_MODE    = "true"
    }
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-ride:2025.04.1-0.8"
  description          = "Studio with custom environment variables"
  is_private           = true
  name                 = "studio-with-env"
  workspace_id         = seqera_workspace.main.id
}
```

### Git Repository

```terraform
# Build a Studio from a Git repository.
#
# Platform reads `.seqera/studio-config.yaml` from the given revision and builds
# the image with Wave on first start, so `data_studio_tool_url` must not be set:
# exactly one of `data_studio_tool_url` or `remote_config` is required.
#
# Anything left unset (cpu, memory, lifespan_hours, ...) is taken from the
# repository's Studio configuration, then from the compute environment or
# workspace defaults. Values set here take precedence over the repository.
resource "seqera_studios" "from_git" {
  name           = "studio-from-git"
  compute_env_id = seqera_compute_env.main.id
  workspace_id   = seqera_workspace.main.id
  auto_start     = false

  remote_config = {
    repository = "https://github.com/my-org/my-studio"
    revision   = "main" # optional; defaults to the repository's default branch
  }

  configuration = {}
}

# Resolved by Platform and available after creation
output "studio_revision" {
  description = "Revision Platform resolved (the default branch when revision is unset)"
  value       = seqera_studios.from_git.remote_config.revision
}

output "studio_commit_id" {
  description = "Commit SHA the Studio was created from"
  value       = seqera_studios.from_git.remote_config.commit_id
}

# Image attributes are read-only. For a Git-based Studio they show what Wave
# built from the repository.
output "studio_base_image" {
  value = seqera_studios.from_git.base_image
}

output "studio_custom_image" {
  description = "True when the Studio runs a custom image built by Wave"
  value       = seqera_studios.from_git.custom_image
}

output "studio_wave_build_url" {
  description = "Wave build for the Studio image, if one was built"
  value       = seqera_studios.from_git.wave_build_url
}
```

### Git Repository Pinned

```terraform
# Build a Studio from a Git repository, pinned to a commit, with some of the
# repository's settings overridden.
#
# `revision` accepts a branch, a tag or a full 40-character commit SHA. Branches
# and tags are resolved when the Studio is created, so the Studio keeps the
# commit it was created from even if the branch later moves. A full SHA pins the
# Studio explicitly. `remote_config.commit_id` is read-only and reports the
# resolved commit.
#
# Changing `remote_config` forces the Studio to be replaced.
resource "seqera_studios" "from_git_pinned" {
  name           = "studio-from-git-pinned"
  description    = "Studio built from a pinned commit"
  compute_env_id = seqera_compute_env.main.id
  workspace_id   = seqera_workspace.main.id
  auto_start     = false

  remote_config = {
    repository = "https://github.com/my-org/my-studio"
    revision   = "3f2b8c1d9e4a7b6c5d0e1f2a3b4c5d6e7f8a9b0c"
  }

  # Anything set here takes precedence over `.seqera/studio-config.yaml`.
  # Anything left unset (cpu, memory, lifespan_hours, ...) is taken from the
  # repository, then from the compute environment or workspace defaults.
  is_private = true

  configuration = {
    cpu            = 4
    memory         = 16384
    lifespan_hours = 12
    environment = {
      STUDIO_ENV = "production"
    }
  }

  label_ids = [seqera_labels.team_datascience.id]
}
```

### Rstudio Mounted Data

```terraform
# Fetch all data links in the workspace
data "seqera_data_links" "workspace_data" {
  workspace_id = seqera_workspace.main.id
}

# Create a lookup map indexed by data link name
locals {
  data_links = {
    for dl in data.seqera_data_links.workspace_data.data_links : dl.name => dl
  }
}

resource "seqera_studios" "rstudio_with_data" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
  configuration = {
    cpu            = 2
    memory         = 8192
    lifespan_hours = 8
    # Mount data links by referencing them by name from the datasource
    # This allows you to dynamically reference S3/Azure/GCS buckets configured in your workspace
    mount_data = [
      local.data_links["my-s3-bucket"].id,
      local.data_links["my-analysis-data"].id,
    ]
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-ride:2025.04.1-0.8"
  description          = "RStudio with mounted S3 data"
  is_private           = true
  name                 = "rstudio-with-data"
  workspace_id         = seqera_workspace.main.id
}
```

### Rstudio Regional Data

```terraform
# Fetch all data links in the workspace
data "seqera_data_links" "workspace_data" {
  workspace_id = seqera_workspace.main.id
}

# Mount only AWS data links in a specific region
resource "seqera_studios" "rstudio_regional_data" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
  configuration = {
    cpu            = 2
    memory         = 8192
    lifespan_hours = 8
    # Filter and mount only AWS data links in us-east-1
    mount_data = [
      for dl in data.seqera_data_links.workspace_data.data_links :
      dl.id if dl.provider == "aws" && dl.region == "us-east-1"
    ]
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-ride:2025.04.1-0.8"
  description          = "RStudio with AWS us-east-1 data only"
  is_private           = true
  name                 = "rstudio-regional-data"
  workspace_id         = seqera_workspace.main.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `compute_env_id` (String) Requires replacement if changed.
- `name` (String) Display name for the Studio session. Requires replacement if changed.
- `workspace_id` (Number) Workspace numeric identifier. Requires replacement if changed.

### Optional

- `allowed_user_ids` (List of Number) IDs of users, besides the creator, allowed to connect to and start this Studio when it is private. Only applies to private Studios; currently limited to a single user. Requires replacement if changed.
- `auto_start` (Boolean) Optionally disable the Studio's automatic launch when it is created. Requires replacement if changed.
- `configuration` (Attributes) Requires replacement if changed. (see [below for nested schema](#nestedatt--configuration))
- `custom_icon_id` (String) Avatar id (from POST /avatars) for the Studio session icon. Semantics on create: null = inherit from the parent checkpoint (if any); empty string = explicit clear (no icon, even if a parent checkpoint has one); non-empty = assign the referenced Avatar. Requires replacement if changed.
- `data_studio_tool_url` (String) Container image URL of the Studio tool. Exactly one of `data_studio_tool_url` or `remote_config` must be set. Requires replacement if changed.
- `description` (String) Description of the Studio session's purpose. Requires replacement if changed.
- `initial_checkpoint_id` (Number) Requires replacement if changed.
- `is_private` (Boolean) Requires replacement if changed.
- `label_ids` (List of Number) List of resource label IDs to associate with this Studio. Reference labels using seqera_labels.label_name.id. Requires replacement if changed.
- `remote_config` (Attributes) Build the Studio from a Git repository. Platform reads `.seqera/studio-config.yaml` from the repository and builds the image with Wave on first start. Values set explicitly on the Studio take precedence over the repository configuration. Exactly one of `data_studio_tool_url` or `remote_config` must be set. Requires replacement if changed. (see [below for nested schema](#nestedatt--remote_config))
- `spot` (Boolean) Whether to use spot or on-demand instances. Studios using Spot instances are not compatible with batch compute environments. Requires replacement if changed.

### Read-Only

- `allowed_users` (Attributes List) (see [below for nested schema](#nestedatt--allowed_users))
- `base_image` (String) Base container image the Studio is built from.
- `custom_icon` (String) Resolved URL for the Studio's user-uploaded icon. Server-derived at read time from the underlying Avatar FK against the current towerPublicEndpoint, so it always reflects the installation's public URL (survives Enterprise domain changes / DB restores). Null when unset — frontends should fall back to the template tool logo.
- `custom_image` (Boolean) Whether the Studio runs a custom image built by Wave.
- `id` (String) Studio session identifier. Same value as `session_id`.
- `session_id` (String) Unique identifier for the Studio session
- `ssh_details` (Attributes) SSH connection details for a Studio session (see [below for nested schema](#nestedatt--ssh_details))
- `wave_build_url` (String) URL of the Wave build for this Studio's image, if one was built.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Optional:

- `conda_environment` (String) Requires replacement if changed.
- `cpu` (Number) Number of CPU cores to allocate. Set to 0 to use the compute environment configured defaults. When omitted, the value from the Git repository's Studio configuration is used if `remote_config` is set, otherwise the compute environment default. Requires replacement if changed.
- `environment` (Map of String) Studio-specific environment variables as key-value pairs. Variable names must contain only alphanumeric and underscore characters, and cannot begin with a number. Requires replacement if changed.
- `gpu` (Number) Set to 0 to disable GPU or 1 to enable GPU. Default: 0; Requires replacement if changed.
- `lifespan_hours` (Number) Maximum lifespan of the Studio session in hours. Requires replacement if changed.
- `memory` (Number) Memory allocation for the Studio session in megabytes (MB). Set to 0 to use the compute environment configured defaults. When omitted, the value from the Git repository's Studio configuration is used if `remote_config` is set, otherwise the compute environment default. Requires replacement if changed.
- `mount_data` (List of String, Deprecated) Requires replacement if changed.
- `mount_data_v2` (Attributes List) Requires replacement if changed. (see [below for nested schema](#nestedatt--configuration--mount_data_v2))
- `ssh_enabled` (Boolean) Requires replacement if changed.

<a id="nestedatt--configuration--mount_data_v2"></a>
### Nested Schema for `configuration.mount_data_v2`

Optional:

- `data_link_id` (String) Requires replacement if changed.
- `path` (String) Requires replacement if changed.



<a id="nestedatt--remote_config"></a>
### Nested Schema for `remote_config`

Optional:

- `repository` (String) URL of the Git repository containing the Studio configuration. Not Null; Requires replacement if changed.
- `revision` (String) Branch, tag or full commit SHA to read the Studio configuration from. When omitted, Platform resolves the repository's default branch and reports the resolved revision. Requires replacement if changed.

Read-Only:

- `commit_id` (String) Commit SHA the Studio was created from. Resolved by Platform from `revision`; to pin a Studio to a commit, set `revision` to the full commit SHA.


<a id="nestedatt--allowed_users"></a>
### Nested Schema for `allowed_users`

Read-Only:

- `avatar` (String)
- `email` (String)
- `id` (Number)
- `user_name` (String)


<a id="nestedatt--ssh_details"></a>
### Nested Schema for `ssh_details`

Read-Only:

- `command` (String) The full SSH command to execute
- `host` (String) The hostname to connect to
- `port` (Number) The SSH port number
- `user` (String) The user in the format 'username@sessionId'

## Import

Import is supported using the following syntax:

In Terraform v1.5.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `id` attribute, for example:

```terraform
import {
  to = seqera_studios.my_seqera_studios
  id = jsonencode({
    id           = "..."
    workspace_id = 0
  })
}
```

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import seqera_studios.my_seqera_studios '{"id": "...", "workspace_id": 0}'
```
