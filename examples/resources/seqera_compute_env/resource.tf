resource "seqera_compute_env" "my_computeenv" {
  config = {
    # ...
  }
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  label_ids = [
    6
  ]
  message      = "...my_message..."
  name         = "...my_name..."
  platform     = "gke-platform"
  workspace_id = 1
}