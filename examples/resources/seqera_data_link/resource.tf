resource "seqera_data_link" "my_datalink" {
  credentials_id    = "...my_credentials_id..."
  description       = "...my_description..."
  name              = "...my_name..."
  provider_type     = "seqeracompute"
  public_accessible = false
  resource_ref      = "...my_resource_ref..."
  type              = "bucket"
  workspace_id      = 4
}