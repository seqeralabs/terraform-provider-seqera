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
  bearer_auth = "xxx=="
}


# data "seqera_user_workspaces" "shahbaz-test-workspace" {
#   user_id = 5
# }