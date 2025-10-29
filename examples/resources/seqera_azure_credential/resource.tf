# Seqera Azure Credentials Examples
#
# Azure credentials store authentication information for accessing Azure services
# within the Seqera Platform workflows. Three authentication modes are supported:
#
# 1. Shared Key - Direct account key authentication (traditional)
# 2. Entra - Azure Entra (formerly Azure AD) service principal authentication
# 3. Cloud - Azure Entra with Cloud-specific configurations
#
# SECURITY BEST PRACTICES:
# - Never hardcode credentials in Terraform files
# - Use Terraform variables marked as sensitive
# - Store actual credentials in secure secret management systems
# - Use service principals (Entra/Cloud) for enhanced security when possible
# - Restrict permissions to minimum required (principle of least privilege)

# Variable declarations for sensitive Azure credentials
variable "azure_batch_key" {
  description = "Azure Batch account key (for shared key auth)"
  type        = string
  sensitive   = true
}

variable "azure_storage_key" {
  description = "Azure Storage account key (for shared key auth)"
  type        = string
  sensitive   = true
}

variable "azure_tenant_id" {
  description = "Azure tenant ID (for Entra/Cloud auth)"
  type        = string
  sensitive   = true
}

variable "azure_client_id" {
  description = "Azure service principal client ID (for Entra/Cloud auth)"
  type        = string
  sensitive   = true
}

variable "azure_client_secret" {
  description = "Azure service principal client secret (for Entra/Cloud auth)"
  type        = string
  sensitive   = true
}

# =============================================================================
# Example 1: Shared Key Authentication (Traditional)
# =============================================================================
# Use this mode when you have direct access to Azure account keys.
# Less secure than service principal authentication.

resource "seqera_azure_credential" "shared_key" {
  name         = "azure-shared-key"
  workspace_id = seqera_workspace.main.id

  # Required for all modes
  batch_name   = "myazurebatch"
  storage_name = "myazurestorage"

  # Shared key mode: Use account keys
  batch_key   = var.azure_batch_key
  storage_key = var.azure_storage_key
}

# =============================================================================
# Example 2: Azure Entra Authentication (Recommended)
# =============================================================================
# Use this mode for modern, identity-based access control.
# More secure than shared keys. Supports Azure RBAC and Managed Identities.

resource "seqera_azure_credential" "entra" {
  name         = "azure-entra"
  workspace_id = seqera_workspace.main.id

  # Required for all modes
  batch_name   = "myazurebatch"
  storage_name = "myazurestorage"

  # Entra mode: Use service principal
  tenant_id     = var.azure_tenant_id
  client_id     = var.azure_client_id
  client_secret = var.azure_client_secret
}

# =============================================================================
# Example 3: Azure Cloud Authentication
# =============================================================================
# Use this mode for Entra identity-based access with Cloud specializations.

resource "seqera_azure_credential" "cloud" {
  name         = "azure-cloud"
  workspace_id = seqera_workspace.main.id

  # Required for all modes
  batch_name   = "myazurebatch"
  storage_name = "myazurestorage"

  # Cloud mode: Use service principal (same as Entra)
  tenant_id     = var.azure_tenant_id
  client_id     = var.azure_client_id
  client_secret = var.azure_client_secret
}

# =============================================================================
# Example 4: Creating Azure Service Principal for Entra/Cloud Auth
# =============================================================================
# To create a service principal for Entra or Cloud authentication:
#
# 1. Create an Azure App Registration:
#    az ad app create --display-name "seqera-platform"
#
# 2. Create a service principal:
#    az ad sp create --id <app-id>
#
# 3. Create a client secret:
#    az ad app credential reset --id <app-id>
#
# 4. Assign necessary permissions:
#    - Batch: Contributor role on Batch account
#    - Storage: Storage Blob Data Contributor on Storage account
#
# 5. Get the tenant ID:
#    az account show --query tenantId -o tsv
#
# 6. Use the App ID as client_id, secret as client_secret, and tenant ID as tenant_id

# =============================================================================
# Example 5: Multiple Environments with Different Auth Modes
# =============================================================================

locals {
  azure_environments = {
    "dev" = {
      mode         = "shared_key"
      batch_name   = "devazurebatch"
      storage_name = "devazurestorage"
    }
    "prod" = {
      mode         = "entra"
      batch_name   = "prodazurebatch"
      storage_name = "prodazurestorage"
    }
  }
}

# Note: This example shows the structure but would need conditional logic
# to handle different authentication modes properly
