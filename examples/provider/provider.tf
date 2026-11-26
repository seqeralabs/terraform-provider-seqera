terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.45.0"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}