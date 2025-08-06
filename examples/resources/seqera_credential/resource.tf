resource "seqera_credential" "my_credential" {
  checked = false
  credentials = {
    base_url    = "https://www.googleapis.com"
    category    = "cloud"
    description = "Google Cloud credentials for production workloads"
    id          = "...my_id..."
    keys = {
      # ...
    }
    name          = "my-gcp-credentials"
    provider_type = "google"
  }
  workspace_id = 6
}