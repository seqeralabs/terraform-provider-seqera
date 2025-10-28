# Seqera Pipeline Secrets Examples
#
# Pipeline secrets store encrypted sensitive data such as API keys, passwords,
# and configuration values that workflows can access during execution.
# Secrets are scoped to a workspace.
#
# SECURITY BEST PRACTICES:
# - Never hardcode secret values in Terraform files
# - Use Terraform variables marked as sensitive
# - Store actual secrets in secure secret management systems
# - Reference secrets using variables or data sources

# Variable declarations for sensitive values
variable "github_token" {
  description = "GitHub API token for workflow access"
  type        = string
  sensitive   = true
}

variable "aws_access_key" {
  description = "AWS access key for S3 access"
  type        = string
  sensitive   = true
}

variable "database_password" {
  description = "Database password for data access"
  type        = string
  sensitive   = true
}

# Example 1: API key secret
# Store a GitHub API token for accessing private repositories

resource "seqera_pipeline_secret" "github_token" {
  name         = "github_api_token"
  value        = var.github_token
  workspace_id = seqera_workspace.main.id
}

# Example 2: Cloud credentials
# Store AWS access credentials for S3 bucket access

resource "seqera_pipeline_secret" "aws_access_key" {
  name         = "aws_access_key_id"
  value        = var.aws_access_key
  workspace_id = seqera_workspace.main.id
}

# Example 3: Database credentials
# Store database password for workflow data access

resource "seqera_pipeline_secret" "db_password" {
  name         = "database_password"
  value        = var.database_password
  workspace_id = seqera_workspace.main.id
}

# Example 4: Multiple secrets for a workspace
# Create several secrets for different services

locals {
  secrets = {
    "slack_webhook_url" = var.slack_webhook
    "dockerhub_token"   = var.dockerhub_token
    "api_endpoint_key"  = var.api_key
  }
}

resource "seqera_pipeline_secret" "service_secrets" {
  for_each = local.secrets

  name         = each.key
  value        = each.value
  workspace_id = seqera_workspace.main.id
}
