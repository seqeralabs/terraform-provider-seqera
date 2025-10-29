# SSH Credentials Example
# These are example non-sensitive values for testing

# SSH credential without passphrase
resource "seqera_ssh_credential" "example_no_passphrase" {
  name        = "Example SSH Credentials (No Passphrase)"
  private_key = <<-EOT
    -----BEGIN OPENSSH PRIVATE KEY-----
    b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABlwAAAAdzc2gtcn
    NhAAAAAwEAAQAAAYEAyExampleKeyContentHereNotARealKey123456789ABCDEFGHIJKLM
    NOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=ExampleKeyContent
    -----END OPENSSH PRIVATE KEY-----
  EOT

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# SSH credential with passphrase
resource "seqera_ssh_credential" "example_with_passphrase" {
  name        = "Example SSH Credentials (With Passphrase)"
  private_key = <<-EOT
    -----BEGIN OPENSSH PRIVATE KEY-----
    b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABB1234567
    ExampleEncryptedKeyContentHereNotARealKey123456789ABCDEFGHIJKLMNOPQRSTUV
    WXYZabcdefghijklmnopqrstuvwxyz0123456789+/=ExampleKeyContent
    -----END OPENSSH PRIVATE KEY-----
  EOT
  passphrase  = "example-passphrase-not-real"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "ssh_credential_id" {
  value       = seqera_ssh_credential.example_no_passphrase.credentials_id
  description = "The ID of the SSH credential"
}

output "ssh_credential_provider_type" {
  value       = seqera_ssh_credential.example_no_passphrase.provider_type
  description = "The provider type (should be 'ssh')"
}
