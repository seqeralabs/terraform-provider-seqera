# Google Cloud Credentials Example
# These are example non-sensitive values for testing

# Google Cloud credential with service account JSON
resource "seqera_google_credential" "example" {
  name = "Example Google Cloud Credentials"

  # This is an example service account key JSON (non-functional)
  data = jsonencode({
    "type" : "service_account",
    "project_id" : "example-project-123456",
    "private_key_id" : "1234567890abcdef1234567890abcdef12345678",
    "private_key" : "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7xYz2LqYqrLYS\nexamplekeycontenthere/notarealkey\n-----END PRIVATE KEY-----\n",
    "client_email" : "example-service-account@example-project-123456.iam.gserviceaccount.com",
    "client_id" : "123456789012345678901",
    "auth_uri" : "https://accounts.google.com/o/oauth2/auth",
    "token_uri" : "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url" : "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url" : "https://www.googleapis.com/robot/v1/metadata/x509/example-service-account%40example-project-123456.iam.gserviceaccount.com",
    "universe_domain" : "googleapis.com"
  })

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "google_credential_id" {
  value       = seqera_google_credential.example.credentials_id
  description = "The ID of the Google Cloud credential"
}

output "google_credential_provider_type" {
  value       = seqera_google_credential.example.provider_type
  description = "The provider type (should be 'google')"
}
