terraform {
  required_providers {
    seqera = {
      source  = "registry.terraform.io/seqeralabs/seqera"
      version = "0.25.1"
    }
  }
  required_version = ">= 1.0"
}

provider "seqera" {
  server_url  = var.seqera_server_url
  bearer_auth = var.seqera_bearer_auth
}