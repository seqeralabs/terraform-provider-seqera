# Seqera Google Cloud Credentials Examples
#
# Google Cloud credentials store service account keys for authenticating with
# Google Cloud Platform services in Seqera workflows.
#
# SECURITY BEST PRACTICES:
# - Never hardcode service account keys in Terraform files
# - Use file() function to read from secure key files
# - Use Terraform variables marked as sensitive
# - Store service account keys in secure secret management systems
# - Restrict service account permissions to minimum required

# Variable declarations for sensitive service account key
variable "gcp_service_account_key" {
  description = "GCP service account key JSON"
  type        = string
  sensitive   = true
}

# Example 1: Basic Google credentials using file
# Load service account key from a JSON file

resource "seqera_google_credential" "gcs_access" {
  name         = "gcs-bucket-access"
  workspace_id = seqera_workspace.main.id

  keys = {
    data = file("${path.module}/gcp-service-account.json")
  }
}

# Example 2: Google credentials using variable
# Use a sensitive variable for the service account key

resource "seqera_google_credential" "gcp_prod" {
  name         = "gcp-production"
  workspace_id = seqera_workspace.prod.id

  keys = {
    data = var.gcp_service_account_key
  }
}

# Example 3: Google credentials with jsonencode
# Build service account key from separate variables

variable "gcp_project_id" {
  type      = string
  sensitive = false
}

variable "gcp_client_email" {
  type      = string
  sensitive = false
}

variable "gcp_private_key" {
  type      = string
  sensitive = true
}

resource "seqera_google_credential" "gcp_composed" {
  name         = "gcp-composed-key"
  workspace_id = seqera_workspace.main.id

  keys = {
    data = jsonencode({
      type                    = "service_account"
      project_id              = var.gcp_project_id
      private_key_id          = "key-id"
      private_key             = var.gcp_private_key
      client_email            = var.gcp_client_email
      client_id               = "123456789"
      auth_uri                = "https://accounts.google.com/o/oauth2/auth"
      token_uri               = "https://oauth2.googleapis.com/token"
      auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    })
  }
}

# Example 4: Multiple GCP credentials for different projects
# Create credentials for multiple GCP projects

locals {
  gcp_projects = {
    "data-pipeline" = "service-account-data-pipeline.json"
    "ml-training"   = "service-account-ml-training.json"
  }
}

resource "seqera_google_credential" "project_credentials" {
  for_each = local.gcp_projects

  name         = "gcp-${each.key}"
  workspace_id = seqera_workspace.main.id

  keys = {
    data = file("${path.module}/${each.value}")
  }
}
