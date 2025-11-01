variable "azure_batch_name" {
  type = string
}

variable "azure_batch_key" {
  type      = string
  sensitive = true
}

variable "azure_storage_name" {
  type = string
}

variable "azure_storage_key" {
  type      = string
  sensitive = true
}

resource "seqera_azure_credential" "example" {
  name         = "azure-main"
  workspace_id = seqera_workspace.main.id

  batch_name   = var.azure_batch_name
  batch_key    = var.azure_batch_key
  storage_name = var.azure_storage_name
  storage_key  = var.azure_storage_key
}
