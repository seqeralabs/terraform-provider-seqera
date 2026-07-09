terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.41.0"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}