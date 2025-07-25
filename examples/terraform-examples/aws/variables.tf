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

# AWS Configuration
variable "aws_region" {
  description = "AWS region for compute environment and resources"
  type        = string
  default     = "us-east-1"

  validation {
    condition = can(regex("^[a-z]{2}-[a-z]+-[0-9]$", var.aws_region))
    error_message = "AWS region must be in the format xx-xxxx-x (e.g., us-east-1)."
  }
}

variable "access_key" {
  description = "AWS access key ID for authentication"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.access_key) > 0
    error_message = "AWS access key cannot be empty."
  }
}

variable "secret_key" {
  description = "AWS secret access key for authentication"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.secret_key) > 0
    error_message = "AWS secret key cannot be empty."
  }
}

variable "iam_role" {
  description = "IAM role ARN to assume for AWS operations (optional)"
  type        = string
  default     = null

  validation {
    condition = var.iam_role == null || can(regex("^arn:aws:iam::[0-9]{12}:role/.+$", var.iam_role))
    error_message = "IAM role must be a valid ARN format (arn:aws:iam::account:role/role-name)."
  }
}

# Workflow Configuration
variable "work_dir" {
  description = "S3 bucket URI for workflow working directory (e.g., s3://my-bucket/work)"
  type        = string

  validation {
    condition     = can(regex("^s3://[a-z0-9.-]+(/.*)?$", var.work_dir))
    error_message = "Work directory must be a valid S3 URI starting with s3://."
  }
}