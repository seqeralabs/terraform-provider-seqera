terraform {
  required_providers {
    seqera = {
      source  = "registry.terraform.io/seqeralabs/seqera"
      version = "0.0.3"
    }
  }
}
provider "seqera" {
  server_url  = var.seqera_server_url
  bearer_auth = var.seqera_bearer_auth
}

variable "seqera_server_url" {
  description = "Seqera API server URL"
  type        = string
  default     = "https://api.cloud.seqera.io"
}

variable "seqera_bearer_auth" {
  description = "Seqera API bearer token"
  type        = string
  sensitive   = true
}

# Test org + workspace. Typed credential resources require a workspace_id;
# the credential blocks below reference seqera_workspace.test.id, giving a
# self-contained plan.
resource "seqera_orgs" "test" {
  name        = "tf-provider-credential-tests"
  full_name   = "Terraform Provider Credential Tests"
  description = "Test organization for the credential resource CI validation"
}

resource "seqera_workspace" "test" {
  name        = "credential-tests"
  full_name   = "Credential Tests"
  description = "Test workspace for credential resource CI validation"
  org_id      = seqera_orgs.test.org_id
  visibility  = "PRIVATE"
}
