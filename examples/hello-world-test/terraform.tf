terraform {
  required_providers {
    seqera = {
      source  = "registry.terraform.io/speakeasy/seqera"
      #version = "0.0.3"
    }
  }
}
provider "seqera" {
  server_url  = var.seqera_server_url
  bearer_auth = var.seqera_bearer_auth
}

variable "seqera_server_url" {
  description = "Seqera API server URL"
  type        = string
  default     = "https://api.cloud.seqera.io"
}

variable "seqera_bearer_auth" {
  description = "Seqera API bearer token"
  type        = string
  sensitive   = true
}