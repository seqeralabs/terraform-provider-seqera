terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.30.4"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}