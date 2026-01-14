terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.26.5"
    }
  }
}

provider "seqera" {
  bearer_auth = "<YOUR_BEARER_AUTH>" # Required
  server_url  = "..."                # Optional
}