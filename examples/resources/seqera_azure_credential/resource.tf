resource "seqera_azure_credential" "my_azurecredential" {
  checked = true
  credentials = {
    base_url       = "...my_base_url..."
    category       = "...my_category..."
    credentials_id = "...my_credentials_id..."
    description    = "...my_description..."
    keys = {
      batch_key    = "YourAzureBatchAccountKeyHere=="
      batch_name   = "myazurebatch"
      storage_key  = "YourAzureStorageAccountKeyHere=="
      storage_name = "myazurestorage"
    }
    name          = "...my_name..."
    provider_type = "azure"
  }
  workspace_id = 0
}