resource "seqera_google_credential" "my_googlecredential" {
  name          = "My Google Credentials"
  description   = "Google Cloud service account for workflow execution"
  provider_type = "google"
  workspace_id  = 123

  # Recommended: Use file() function to read the service account key JSON
  keys = {
    data = file("${path.module}/service-account-key.json")
  }

  # Alternative: Inline JSON (not recommended for production)
  # keys = {
  #   data = jsonencode({
  #     type         = "service_account"
  #     project_id   = "my-project"
  #     private_key_id = "key-id"
  #     private_key  = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  #     client_email = "service-account@my-project.iam.gserviceaccount.com"
  #     client_id    = "123456789"
  #     auth_uri     = "https://accounts.google.com/o/oauth2/auth"
  #     token_uri    = "https://oauth2.googleapis.com/token"
  #   })
  # }
}