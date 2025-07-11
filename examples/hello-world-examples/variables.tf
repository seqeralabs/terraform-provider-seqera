locals {
  service_account_key = file("${path.module}/service-account-key.json")
  gcp_work_dir        = "gs://terraform-provider-testing"
  azure_batch_name    = "seqeralabs"
  azure_storage_name  = "seqeralabs"
  azure_work_dir      = "az://terraform-provider"
}


variable "access_key" {
  description = "Access key for AWS"
  type        = string

}

variable "secret_key" {
  description = "Secret key for AWS Batch"
  type        = string
  sensitive   = true

}

variable "azure_batch_key" {
  type        = string
  description = "Azyre batch access key"
  sensitive   = true
}


variable "azure_storage_key" {
  type        = string
  description = "Azure storage access key"
  sensitive   = true
}
