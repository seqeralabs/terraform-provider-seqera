resource "seqera_credential" "my_credential" {
  base_url    = "https://www.googleapis.com"
  category    = "cloud"
  checked     = false
  description = "Google Cloud credentials for production workloads"
  id          = "...my_id..."
  keys = {
    google = {
      data          = "{\n  \"type\": \"service_account\",\n  \"project_id\": \"my-project\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account@my-project.iam.gserviceaccount.com\",\n  \"client_id\": \"123456789\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://oauth2.googleapis.com/token\"\n}\n"
      discriminator = "...my_discriminator..."
    }
  }
  name          = "my-gcp-credentials"
  provider_type = "google"
  workspace_id  = 6
}