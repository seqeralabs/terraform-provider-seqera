terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.40.0-RC4"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}