# Google Cloud (GCP) Credential Resource Examples
#
# IMPORTANT: Credential names must use only letters, numbers, underscores, and hyphens.
# No spaces allowed. Use snake_case (my_credential) or kebab-case (my-credential).

# Variable for sensitive credentials
variable "gcp_service_account_key" {
  description = "GCP service account key JSON (as string)"
  type        = string
  sensitive   = true
}

# Example: Basic GCP credentials using service account key
resource "seqera_google_credential" "example" {
  name         = "gcp-main"
  workspace_id = seqera_workspace.main.id

  key = var.gcp_service_account_key
}
