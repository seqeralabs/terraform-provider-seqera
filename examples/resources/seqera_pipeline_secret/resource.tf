variable "github_token" {
  description = "GitHub API token for workflow access"
  type        = string
  sensitive   = true
}

resource "seqera_pipeline_secret" "github_token" {
  name         = "github_api_token"
  value        = var.github_token
  workspace_id = seqera_workspace.main.id
}
