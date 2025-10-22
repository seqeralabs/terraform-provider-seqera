resource "seqera_credential" "my_credential" {
  base_url       = "https://www.googleapis.com"
  category       = "cloud"
  checked        = false
  credentials_id = "...my_credentials_id.."
  description    = "Google Cloud credentials for production workloads"
  keys = {
    local = {
      password = "...my_password..."
    }
  }
  name          = "my-gcp-credentials"
  provider_type = "google"
  workspace_id  = 6
}