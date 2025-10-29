# Azure Credentials Example
# These are example non-sensitive values for testing

# Azure credential with shared keys
resource "seqera_azure_credential" "example_shared_key" {
  name         = "Example Azure Credentials (Shared Key)"
  batch_name   = "examplebatchaccount"
  storage_name = "examplestorageaccount"
  batch_key    = "exampleBatchKeyBase64EncodedString123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ=="
  storage_key  = "exampleStorageKeyBase64EncodedString123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ=="

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Azure credential with service principal (Entra ID)
resource "seqera_azure_credential" "example_service_principal" {
  name          = "Example Azure Credentials (Service Principal)"
  batch_name    = "examplebatchaccount"
  storage_name  = "examplestorageaccount"
  tenant_id     = "12345678-1234-1234-1234-123456789012"
  client_id     = "87654321-4321-4321-4321-210987654321"
  client_secret = "example.client.secret~123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "azure_credential_id" {
  value       = seqera_azure_credential.example_shared_key.credentials_id
  description = "The ID of the Azure credential"
}

output "azure_credential_provider_type" {
  value       = seqera_azure_credential.example_shared_key.provider_type
  description = "The provider type (should be 'azure')"
}
