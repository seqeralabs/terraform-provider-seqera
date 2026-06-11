terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.40.0"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}