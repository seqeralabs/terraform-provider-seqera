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

# GCP Configuration
variable "gcp_region" {
  description = "GCP region for compute environment and resources"
  type        = string
  default     = "us-central1"

  validation {
    condition = can(regex("^[a-z]+-[a-z]+[0-9]$", var.gcp_region))
    error_message = "GCP region must be in the format region-zone (e.g., us-central1, europe-west1)."
  }
}

variable "gcp_location" {
  description = "GCP location for Batch operations (typically same as region)"
  type        = string
  default     = "us-central1"

  validation {
    condition = can(regex("^[a-z]+-[a-z]+[0-9](-[a-z])?$", var.gcp_location))
    error_message = "GCP location must be a valid region or zone (e.g., us-central1, us-central1-a)."
  }
}

variable "service_account_key" {
  description = "File path to the GCP service account key JSON file"
  type        = string

  validation {
    condition     = length(var.service_account_key) > 0
    error_message = "Service account key file path cannot be empty."
  }
}

# Workflow Configuration
variable "work_dir" {
  description = "Google Cloud Storage URI for workflow working directory (e.g., gs://bucket-name/work)"
  type        = string

  validation {
    condition     = can(regex("^gs://[a-z0-9._-]+(/.*)?$", var.work_dir))
    error_message = "Work directory must be a valid GCS URI starting with gs://."
  }
}