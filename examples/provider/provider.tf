terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.40.0-RC3"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}