variable "access_key" {
    description = "Access key for AWS"
    type= string
  
}

variable "secret_key" {
    description = "Secret key for AWS Batch"
    type = string 
    sensitive = true
  
}

locals {
  service_account_key = file("${path.module}/service-account-key.json")
  gcp_work_dir = "gs://terraform-provider-testing"
}
