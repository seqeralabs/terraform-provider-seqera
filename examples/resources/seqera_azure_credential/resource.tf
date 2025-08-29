resource "seqera_azure_credential" "my_azurecredential" {
  base_url       = "...my_base_url..."
  category       = "...my_category..."
  checked        = true
  credentials_id = "...my_credentials_id..."
  date_created   = "2020-01-26T14:24:16.244Z"
  deleted        = true
  description    = "...my_description..."
  keys = {
    batch_key    = "YourAzureBatchAccountKeyHere=="
    batch_name   = "myazurebatch"
    storage_key  = "YourAzureStorageAccountKeyHere=="
    storage_name = "myazurestorage"
  }
  last_updated  = "2022-05-23T22:58:45.279Z"
  last_used     = "2022-09-15T14:55:29.777Z"
  name          = "...my_name..."
  provider_type = "azure"
  workspace_id  = 0
}