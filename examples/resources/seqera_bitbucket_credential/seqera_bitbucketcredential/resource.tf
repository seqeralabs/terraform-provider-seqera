# Seqera BitBucket Credentials Examples
#
# BitBucket credentials store authentication information for accessing BitBucket
# repositories within the Seqera Platform workflows.
#
# SECURITY BEST PRACTICES:
# - Never hardcode credentials in Terraform files
# - Use Terraform variables marked as sensitive
# - Store actual credentials in secure secret management systems
# - Use API tokens instead of app passwords (app passwords are deprecated)
# - Restrict token permissions to minimum required (principle of least privilege)

# Variable declarations for sensitive BitBucket credentials
variable "bitbucket_username" {
  description = "BitBucket account username or email"
  type        = string
  sensitive   = true
}

variable "bitbucket_api_token" {
  description = "BitBucket API token (not app password)"
  type        = string
  sensitive   = true
}

# =============================================================================
# Example 1: Basic BitBucket Credentials with API Token (Recommended)
# =============================================================================
# Use API tokens for authentication. App passwords are deprecated and will
# be removed in June 2026.

resource "seqera_bitbucket_credential" "basic" {
  name         = "bitbucket-main"
  workspace_id = seqera_workspace.main.id

  username = var.bitbucket_username
  token    = var.bitbucket_api_token
}

# =============================================================================
# Example 2: BitBucket Credentials with On-Premises Server
# =============================================================================
# Use base_url when connecting to an on-premises BitBucket server or
# to associate credentials with a specific repository URL.

resource "seqera_bitbucket_credential" "on_prem" {
  name         = "bitbucket-enterprise"
  workspace_id = seqera_workspace.main.id

  username = var.bitbucket_username
  token    = var.bitbucket_api_token
  base_url = "https://bitbucket.mycompany.com/myorg"
}

# =============================================================================
# Example 3: Multiple BitBucket Credentials for Different Organizations
# =============================================================================

locals {
  bitbucket_orgs = {
    "data-science" = {
      username = var.bitbucket_ds_username
      token    = var.bitbucket_ds_token
      base_url = "https://bitbucket.org/data-science-org"
    }
    "ml-research" = {
      username = var.bitbucket_ml_username
      token    = var.bitbucket_ml_token
      base_url = "https://bitbucket.org/ml-research-org"
    }
  }
}

# Note: You would need to declare the corresponding variables:
# variable "bitbucket_ds_username" { type = string; sensitive = true }
# variable "bitbucket_ds_token" { type = string; sensitive = true }
# variable "bitbucket_ml_username" { type = string; sensitive = true }
# variable "bitbucket_ml_token" { type = string; sensitive = true }

resource "seqera_bitbucket_credential" "org_credentials" {
  for_each = local.bitbucket_orgs

  name         = "bitbucket-${each.key}"
  workspace_id = seqera_workspace.main.id
  username     = each.value.username
  token        = each.value.token
  base_url     = each.value.base_url
}

# =============================================================================
# Example 4: Creating BitBucket API Tokens
# =============================================================================
# To create a BitBucket API token (recommended over app passwords):
#
# 1. Go to BitBucket Settings:
#    https://bitbucket.org/account/settings/
#
# 2. Navigate to "Personal access tokens" or "App passwords" (deprecated)
#
# 3. For API Tokens (recommended):
#    - Click "Create token"
#    - Select required permissions (typically: repository:read, repository:write)
#    - Copy the generated token (shown only once)
#
# 4. For App Passwords (deprecated, will be removed June 2026):
#    - Click "Create app password"
#    - Select permissions
#    - Copy the generated password
#
# 5. Use the token/password in your Terraform configuration:
#    - Username: Your BitBucket username or email
#    - Token: The API token or app password

# =============================================================================
# Example 5: Using BitBucket Credentials with Pipelines
# =============================================================================

resource "seqera_bitbucket_credential" "pipeline_creds" {
  name         = "bitbucket-pipelines"
  workspace_id = seqera_workspace.main.id
  username     = var.bitbucket_username
  token        = var.bitbucket_api_token
}

resource "seqera_pipeline" "from_bitbucket" {
  name         = "my-pipeline"
  workspace_id = seqera_workspace.main.id

  # Reference the BitBucket repository
  repository   = "https://bitbucket.org/myorg/my-repo"

  # Use the BitBucket credentials
  credentials_id = seqera_bitbucket_credential.pipeline_creds.credentials_id
}
