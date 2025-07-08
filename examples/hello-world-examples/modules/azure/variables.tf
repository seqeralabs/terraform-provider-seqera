variable "workspace_id" {
    description = "The ID of the workspace where the resources will be created."
    type        = string
}

variable "batch_key" {
    description = "Azure Batch key"
    type = string
    sensitive = true
}

variable "batch_name" {
    description = "Azure Batch name"
    type = string
}

variable "storage_key" {
    description = "Azure storage key"
    type = string
    sensitive = true
}

variable "storage_name" {
    description = "Azure storage name"
    type = string
}

variable "work_dir" {
    description = ""
    type = string
}
