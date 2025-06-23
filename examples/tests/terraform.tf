terraform {
  required_providers {
    seqera = {
      source  = "registry.terraform.io/speakeasy/seqera"
      #version = "0.0.3"
    }
  }
}

provider "seqera" {
  server_url = "https://api.cloud.seqera.io"
  bearer_auth = "eyJ0aWQiOiAxMTkwNH0uOGI1ZGJmNDViMDg5MDYxMjYwNGU2OTZiZTRkYjUzMGYzMGNjNWU5Yg=="
}
