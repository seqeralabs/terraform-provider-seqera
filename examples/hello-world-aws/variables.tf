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
