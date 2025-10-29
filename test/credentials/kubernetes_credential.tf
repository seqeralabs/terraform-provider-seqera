# Kubernetes Credentials Example
# These are example non-sensitive values for testing

# Kubernetes credential with service account token
resource "seqera_kubernetes_credential" "example_token" {
  name  = "Example Kubernetes Credentials (Token)"
  token = "eyJhbGciOiJSUzI1NiIsImtpZCI6IkV4YW1wbGVLZXlJZDEyMzQ1Njc4OTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImV4YW1wbGUtc2VjcmV0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImV4YW1wbGUtc2EiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIxMjM0NTY3OC0xMjM0LTEyMzQtMTIzNC0xMjM0NTY3ODkwMTIiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpleGFtcGxlLXNhIn0.ExampleSignatureNotRealJustForTestingPurposes"

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Kubernetes credential with client certificate
resource "seqera_kubernetes_credential" "example_certificate" {
  name = "Example Kubernetes Credentials (Certificate)"
  client_certificate = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIDITCCAgmgAwIBAgIIExampleCertificateNotRealJustForTesting123456789AB
    CDEFGHIJKLMNOPQRSTUVWXYZ0123456789ExampleCertificateContent
    -----END CERTIFICATE-----
  EOT
  private_key = <<-EOT
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEAyExamplePrivateKeyNotRealJustForTesting123456789ABCDEF
    GHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789ExampleKey
    -----END RSA PRIVATE KEY-----
  EOT

  # Optional: Associate with a workspace
  # workspace_id = data.seqera_workspace.example.id
}

# Output the credential ID
output "kubernetes_credential_id" {
  value       = seqera_kubernetes_credential.example_token.credentials_id
  description = "The ID of the Kubernetes credential"
}

output "kubernetes_credential_provider_type" {
  value       = seqera_kubernetes_credential.example_token.provider_type
  description = "The provider type (should be 'k8s')"
}
