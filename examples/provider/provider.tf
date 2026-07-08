terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.40.3"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}