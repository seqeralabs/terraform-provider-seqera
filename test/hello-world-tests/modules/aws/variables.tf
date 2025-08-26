variable "workspace_id" {
  description = "The ID of the workspace where the resources will be created."
  type        = string
}

# TODO: Export the below as TF_VARs for now, will pull from secrets manager later
#  export TF_VAR_secret_key="xxxx"
#  export TF_VAR_access_key="xxxxx
#
variable "secret_key" {
  type        = string
  description = "AWS secret key for the credential"
  sensitive   = true
}

variable "access_key" {
  type        = string
  description = "AWS access key for the credential"
  sensitive   = true

}

variable "work_dir" {
  description = "S3 working directory for pipelines"
  type        = string
}

variable "iam_role" {
  description = "IAM Role to be used by the platform credential"
  type        = string
}

# Azure credential variables
variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
  default     = ""
}

variable "azure_tenant_id" {
  description = "Azure tenant ID"
  type        = string
  default     = ""
}

variable "azure_client_id" {
  description = "Azure client ID"
  type        = string
  default     = ""
}

variable "azure_client_secret" {
  description = "Azure client secret"
  type        = string
  sensitive   = true
  default     = ""
}

# Google credential variables
variable "google_project_id" {
  description = "Google Cloud project ID"
  type        = string
  default     = ""
}

variable "google_service_account_key" {
  description = "Google service account key (JSON)"
  type        = string
  sensitive   = true
  default     = ""
}
