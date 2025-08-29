resource "seqera_google_credential" "my_googlecredential" {
  base_url       = "...my_base_url..."
  category       = "...my_category..."
  checked        = false
  credentials_id = "...my_credentials_id..."
  date_created   = "2022-11-20T02:44:17.334Z"
  deleted        = false
  description    = "...my_description..."
  keys = {
    data = "{\n  \"type\": \"service_account\",\n  \"project_id\": \"my-project\",\n  \"private_key_id\": \"key-id\",\n  \"private_key\": \"-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----\\n\",\n  \"client_email\": \"service-account@my-project.iam.gserviceaccount.com\",\n  \"client_id\": \"123456789\",\n  \"auth_uri\": \"https://accounts.google.com/o/oauth2/auth\",\n  \"token_uri\": \"https://oauth2.googleapis.com/token\"\n}\n"
  }
  last_updated  = "2022-05-09T00:51:23.496Z"
  last_used     = "2021-10-25T20:07:00.111Z"
  name          = "...my_name..."
  provider_type = "google"
  workspace_id  = 6
}