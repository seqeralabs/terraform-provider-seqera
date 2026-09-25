terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.43.0-RC2"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}