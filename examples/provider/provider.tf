terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.27.0"
    }
  }
}

provider "seqera" {
  bearer_auth = "<YOUR_BEARER_AUTH>" # Required
  server_url  = "..."                # Optional
}