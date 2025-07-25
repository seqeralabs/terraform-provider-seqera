# Seqera Platform Configuration
variable "seqera_server_url" {
  description = "Seqera Platform API server URL"
  type        = string
  default     = "https://api.cloud.seqera.io"

  validation {
    condition = can(regex("^https?://", var.seqera_server_url))
    error_message = "Server URL must be a valid HTTP or HTTPS URL."
  }
}

variable "seqera_bearer_auth" {
  description = "Seqera Platform API bearer token for authentication"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.seqera_bearer_auth) > 0
    error_message = "Bearer token cannot be empty."
  }
}

# Azure Configuration
variable "azure_region" {
  description = "Azure region for compute environment and resources"
  type        = string
  default     = "eastus"

  validation {
    condition = can(regex("^[a-z0-9]+$", var.azure_region))
    error_message = "Azure region must be a valid region name (e.g., eastus, westus2)."
  }
}

variable "batch_name" {
  description = "Azure Batch account name"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]{3,24}$", var.batch_name))
    error_message = "Batch account name must be 3-24 characters long and contain only lowercase letters and numbers."
  }
}

variable "batch_key" {
  description = "Azure Batch account access key"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.batch_key) > 0
    error_message = "Batch key cannot be empty."
  }
}

variable "storage_name" {
  description = "Azure Storage account name"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]{3,24}$", var.storage_name))
    error_message = "Storage account name must be 3-24 characters long and contain only lowercase letters and numbers."
  }
}

variable "storage_key" {
  description = "Azure Storage account access key"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.storage_key) > 0
    error_message = "Storage key cannot be empty."
  }
}

# Workflow Configuration
variable "work_dir" {
  description = "Azure Blob storage URI for workflow working directory (e.g., az://container/path)"
  type        = string

  validation {
    condition     = can(regex("^az://[a-z0-9-]+(/.*)?$", var.work_dir))
    error_message = "Work directory must be a valid Azure Blob URI starting with az://."
  }
}