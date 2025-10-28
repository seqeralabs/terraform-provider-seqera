# Seqera Gitea Credentials Examples
#
# Gitea credentials store authentication information for accessing Gitea
# repositories within the Seqera Platform workflows.
#
# SECURITY BEST PRACTICES:
# - Never hardcode credentials in Terraform files
# - Use Terraform variables marked as sensitive
# - Store actual credentials in secure secret management systems
# - Regularly rotate passwords
# - Use strong passwords with appropriate complexity

# Variable declarations for sensitive Gitea credentials
variable "gitea_username" {
  description = "Gitea account username"
  type        = string
  sensitive   = true
}

variable "gitea_password" {
  description = "Gitea account password"
  type        = string
  sensitive   = true
}

# =============================================================================
# Example 1: Basic Gitea Credentials
# =============================================================================
# Basic configuration with Gitea username and password.

resource "seqera_gitea_credential" "basic" {
  name         = "gitea-main"
  workspace_id = seqera_workspace.main.id

  username = var.gitea_username
  password = var.gitea_password
}

# =============================================================================
# Example 2: Gitea Credentials with Self-Hosted Server
# =============================================================================
# Use base_url when connecting to a self-hosted Gitea server or
# to associate credentials with a specific repository URL.

resource "seqera_gitea_credential" "self_hosted" {
  name         = "gitea-enterprise"
  workspace_id = seqera_workspace.main.id

  username = var.gitea_username
  password = var.gitea_password
  base_url = "https://gitea.mycompany.com"
}

# =============================================================================
# Example 3: Multiple Gitea Credentials for Different Servers
# =============================================================================

locals {
  gitea_servers = {
    "public" = {
      username = var.gitea_public_username
      password = var.gitea_public_password
      base_url = "https://gitea.io"
    }
    "internal" = {
      username = var.gitea_internal_username
      password = var.gitea_internal_password
      base_url = "https://gitea.internal.company.com"
    }
    "dev" = {
      username = var.gitea_dev_username
      password = var.gitea_dev_password
      base_url = "https://gitea-dev.company.com"
    }
  }
}

# Note: You would need to declare the corresponding variables:
# variable "gitea_public_username" { type = string; sensitive = true }
# variable "gitea_public_password" { type = string; sensitive = true }
# variable "gitea_internal_username" { type = string; sensitive = true }
# variable "gitea_internal_password" { type = string; sensitive = true }
# variable "gitea_dev_username" { type = string; sensitive = true }
# variable "gitea_dev_password" { type = string; sensitive = true }

resource "seqera_gitea_credential" "multi_server" {
  for_each = local.gitea_servers

  name         = "gitea-${each.key}"
  workspace_id = seqera_workspace.main.id
  username     = each.value.username
  password     = each.value.password
  base_url     = each.value.base_url
}

# =============================================================================
# Example 4: Gitea Credentials with Specific Repository URL
# =============================================================================
# Associate credentials with a specific repository URL for better organization.

resource "seqera_gitea_credential" "with_repo_url" {
  name         = "gitea-project-repo"
  workspace_id = seqera_workspace.main.id

  username = var.gitea_username
  password = var.gitea_password
  base_url = "https://try.gitea.io/seqera/tower"
}

# =============================================================================
# Example 5: Using Gitea Credentials with Pipelines
# =============================================================================

resource "seqera_gitea_credential" "pipeline_creds" {
  name         = "gitea-pipelines"
  workspace_id = seqera_workspace.main.id
  username     = var.gitea_username
  password     = var.gitea_password
  base_url     = "https://gitea.mycompany.com"
}

resource "seqera_pipeline" "from_gitea" {
  name         = "my-pipeline"
  workspace_id = seqera_workspace.main.id

  # Reference the Gitea repository
  repository = "https://gitea.mycompany.com/myorg/my-repo"

  # Use the Gitea credentials
  credentials_id = seqera_gitea_credential.pipeline_creds.credentials_id
}

# =============================================================================
# Example 6: Gitea Credentials for Public and Private Repositories
# =============================================================================
# Different credentials may be needed for public vs private repositories
# or different organizations.

resource "seqera_gitea_credential" "public_repos" {
  name         = "gitea-public"
  workspace_id = seqera_workspace.main.id
  username     = var.gitea_public_username
  password     = var.gitea_public_password
}

resource "seqera_gitea_credential" "private_repos" {
  name         = "gitea-private"
  workspace_id = seqera_workspace.main.id
  username     = var.gitea_private_username
  password     = var.gitea_private_password
  base_url     = "https://gitea.internal.company.com"
}

# =============================================================================
# Example 7: Setting Up Gitea Authentication
# =============================================================================
# To create credentials for Gitea:
#
# 1. Log into your Gitea server:
#    https://gitea.example.com
#
# 2. Go to Settings > Account > Password
#    - Ensure you have a strong password
#
# 3. For API access, consider creating an Access Token instead:
#    - Settings > Applications > Generate New Token
#    - Select appropriate scopes (typically: repo, read:user)
#    - Note: Access tokens provide better security than passwords
#
# 4. Use the credentials in your Terraform configuration:
#    - Username: Your Gitea username
#    - Password: Your Gitea password or access token
#
# SECURITY NOTE:
# - Access tokens are recommended over passwords for API access
# - Tokens can be revoked individually without changing your password
# - Consider using separate tokens for different services/purposes

# =============================================================================
# Example 8: Gitea Credentials with Organization Context
# =============================================================================
# When working with multiple organizations, organize credentials by org.

locals {
  gitea_organizations = {
    "data-science" = "gitea-ds"
    "ml-research"  = "gitea-ml"
    "production"   = "gitea-prod"
  }
}

resource "seqera_gitea_credential" "org_credentials" {
  for_each = local.gitea_organizations

  name         = each.value
  workspace_id = seqera_workspace.main.id
  username     = var.gitea_username
  password     = var.gitea_password
  base_url     = "https://gitea.mycompany.com/${each.key}"
}
