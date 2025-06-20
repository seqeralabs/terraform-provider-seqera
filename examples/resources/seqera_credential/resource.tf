resource "seqera_credential" "my_credential" {
  base_url       = "...my_base_url..."
  category       = "...my_category..."
  checked        = false
  credentials_id = "...my_credentials_id.."
  description    = "...my_description..."
  keys = {
    google = {
      data          = "...my_data..."
      discriminator = "...my_discriminator..."
    }
  }
  name          = "...my_name..."
  provider_type = "...my_provider_t"
  workspace_id  = 6
}