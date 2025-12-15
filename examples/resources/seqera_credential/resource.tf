resource "seqera_credential" "my_credential" {
  base_url    = "https://www.googleapis.com"
  category    = "cloud"
  checked     = false
  description = "Google Cloud credentials for production workloads"
  id          = "...my_id..."
  keys = {
    local = {
      credential_type = "...my_credential_type..."
      discriminator   = "...my_discriminator..."
      password        = "...my_password..."
    }
  }
  name          = "my-gcp-credentials"
  provider_type = "google"
  workspace_id  = 6
}