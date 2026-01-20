resource "seqera_credential" "my_credential" {
  base_url    = "https://www.googleapis.com"
  category    = "cloud"
  checked     = false
  description = "Google Cloud credentials for production workloads"
  id          = "...my_id..."
  keys = {
    s3 = {
      access_key                = "...my_access_key..."
      path_style_access_enabled = false
      secret_key                = "...my_secret_key..."
    }
  }
  name          = "my-gcp-credentials"
  provider_type = "google"
  workspace_id  = 6
}