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
    type = string
}

variable "iam_role" {
    description = "IAM Role to be used by the platform credential"
    type = string
}