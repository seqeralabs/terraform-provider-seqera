terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.30.1"
    }
  }
}

provider "seqera" {
  bearer_auth = "<YOUR_BEARER_AUTH>" # Required
  server_url  = "..."                # Optional
}