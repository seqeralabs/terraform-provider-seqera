resource "seqera_compute_env" "my_computeenv" {
  compute_env = {
    config = {
      # ...
    }
    credentials_id = "...my_credentials_id..."
    description    = "...my_description..."
    message        = "...my_message..."
    name           = "...my_name..."
    platform       = "google-lifesciences"
  }
  label_ids = [
    6
  ]
  workspace_id = 1
}