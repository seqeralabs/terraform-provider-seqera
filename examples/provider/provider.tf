terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.42.0"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}