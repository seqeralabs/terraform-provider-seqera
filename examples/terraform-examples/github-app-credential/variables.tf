variable "workspace_id" {
  description = "Numeric ID of the existing Seqera workspace that will own the credential."
  type        = number
}

variable "credential_name" {
  description = "Name for the credential in Seqera."
  type        = string
  default     = "github-app"

  validation {
    condition     = can(regex("^[a-zA-Z0-9_-]{2,99}$", var.credential_name))
    error_message = "credential_name must contain 2-99 letters, numbers, underscores, or hyphens."
  }
}

variable "github_app_id" {
  description = "GitHub App ID from the app's General settings page."
  type        = string
}

variable "github_app_client_id" {
  description = "GitHub App client ID from the app's General settings page."
  type        = string
}

variable "github_app_private_key_path" {
  description = "Local path to the GitHub App PEM private key. Do not commit this file."
  type        = string
  sensitive   = true
}

variable "github_base_url" {
  description = "Repository URL scope for this credential."
  type        = string
  default     = null
  nullable    = true
}

variable "seqera_server_url" {
  description = "Optional Seqera API URL. Defaults to Seqera Cloud."
  type        = string
  default     = null
  nullable    = true
}
