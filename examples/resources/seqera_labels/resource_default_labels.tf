locals {
  default_labels = {
    "team"        = "data-science"
    "cost-center" = "research"
    "project"     = "genomics"
  }
}

resource "seqera_labels" "defaults" {
  for_each = local.default_labels

  workspace_id = seqera_workspace.main.id
  name         = each.key
  value        = each.value
  resource     = true
  is_default   = true
}
