variable "workspace_id" {
    description = "The ID of the workspace where the resources will be created."
    type        = string
}


variable "work_dir" {
    description = "Working directory for nextflow runs"
    type = string
}

variable "service_account_key" {
    description = "Service account json key"
    type = string
}