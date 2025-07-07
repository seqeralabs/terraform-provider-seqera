variable "access_key" {
    description = "Access key for AWS"
    type= string
  
}

variable "secret_key" {
    description = "Secret key for AWS Batch"
    type = string 
    sensitive = true
  
}