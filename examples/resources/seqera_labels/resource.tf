# Example 1: Resource Label (Most Common)
# Resource labels have key-value pairs used to tag pipelines, workflows, etc.
# The "value" field is REQUIRED when resource=true.

resource "seqera_labels" "environment" {
  workspace_id = 123456
  name         = "environment"
  value        = "production"
  resource     = true
  is_default   = false
}

# Example 2: Default Resource Labels (Auto-Applied)
# Default labels are automatically applied to new resources in the workspace.
# Use for_each to create multiple default labels from a map.

locals {
  default_labels = {
    "team"        = "data-science"
    "cost-center" = "research"
    "project"     = "genomics"
  }
}

resource "seqera_labels" "defaults" {
  for_each = local.default_labels

  workspace_id = 123456
  name         = each.key
  value        = each.value
  resource     = true
  is_default   = true
}

# Example 3: Non-Resource Labels (Simple Tags)
# Non-resource labels are simple tags without values.
# These cannot have a value and cannot be marked as default.

resource "seqera_labels" "critical" {
  workspace_id = 123456
  name         = "critical"
  resource     = false
  is_default   = false
}

resource "seqera_labels" "experimental" {
  workspace_id = 123456
  name         = "experimental"
  resource     = false
  is_default   = false
}

# Example 4: Labels with Workspace Reference
# Use workspace resource references for dynamic workspace IDs.

resource "seqera_workspace" "my_workspace" {
  name          = "my-workspace"
  org_id        = 123456
  full_name     = "my-org/my-workspace"
  visibility    = "PRIVATE"
  description   = "Example workspace"
}

resource "seqera_labels" "workspace_label" {
  workspace_id = seqera_workspace.my_workspace.id
  name         = "owner"
  value        = "john-doe"
  resource     = true
  is_default   = false
}
