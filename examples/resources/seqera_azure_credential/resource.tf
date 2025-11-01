# Azure Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variables for sensitive credentials
variable "azure_batch_name" {
  description = "Azure Batch account name"
  type        = string
}

variable "azure_batch_key" {
  description = "Azure Batch account key"
  type        = string
  sensitive   = true
}

variable "azure_storage_name" {
  description = "Azure Storage account name"
  type        = string
}

variable "azure_storage_key" {
  description = "Azure Storage account key"
  type        = string
  sensitive   = true
}

# Example: Basic Azure credentials
resource "seqera_azure_credential" "example" {
  name         = "azure-main"
  workspace_id = seqera_workspace.main.id

  batch_name    = var.azure_batch_name
  batch_key     = var.azure_batch_key
  storage_name  = var.azure_storage_name
  storage_key   = var.azure_storage_key
}
