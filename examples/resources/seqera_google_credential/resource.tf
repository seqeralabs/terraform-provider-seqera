resource "seqera_google_credential" "my_googlecredential" {
  base_url       = "...my_base_url..."
  category       = "...my_category..."
  checked        = false
  credentials_id = "...my_credentials_id..."
  description    = "Google Cloud service account for workflow execution"
  keys = {
    data = "file(\"${path.module}/service-account-key.json\")"
  }
  name          = "My Google Cloud Credentials"
  provider_type = "google"
  workspace_id  = 123
}