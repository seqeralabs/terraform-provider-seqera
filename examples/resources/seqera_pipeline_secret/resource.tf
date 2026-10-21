variable "github_token" {
  description = "GitHub API token for workflow access"
  type        = string
  sensitive   = true
}

# `value` is write-only (Terraform >= 1.11): it is sent to the API but never
# stored in state, so it can come from a variable or an ephemeral resource.
resource "seqera_pipeline_secret" "github_token" {
  name         = "github_api_token"
  value        = var.github_token
  workspace_id = seqera_workspace.main.id
}

# Rotating a secret: Terraform cannot detect changes to a write-only argument,
# so bump `value_version` together with `value` to update the secret in place.
resource "seqera_pipeline_secret" "rotated" {
  name          = "database_password"
  value         = var.database_password
  value_version = 2 # increment whenever `value` changes
  workspace_id  = seqera_workspace.main.id
}
