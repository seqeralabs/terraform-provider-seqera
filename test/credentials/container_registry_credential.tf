# Container Registry Credentials Example
# These are example non-sensitive values for testing

# Docker Hub credential
resource "seqera_container_registry_credential" "example_docker_hub" {
  name      = "example-docker-hub-credentials"
  user_name = "example-docker-user"
  password  = "example-docker-password-123456"
  registry  = "docker.io"

  workspace_id = var.seqera_workspace_id
}

# AWS ECR credential
resource "seqera_container_registry_credential" "example_ecr" {
  name      = "example-aws-ecr-credentials"
  user_name = "AWS"
  password  = "example-ecr-token-base64-encoded-string-123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  registry  = "123456789012.dkr.ecr.us-east-1.amazonaws.com"

  workspace_id = var.seqera_workspace_id
}

# Google Container Registry (GCR) credential
resource "seqera_container_registry_credential" "example_gcr" {
  name      = "example-gcr-credentials"
  user_name = "_json_key"
  password = jsonencode({
    "type" : "service_account",
    "project_id" : "example-project-123456",
    "private_key_id" : "1234567890abcdef1234567890abcdef12345678",
    "private_key" : "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7xYz2LqYqrLYS\nexamplekeycontenthere/notarealkey\n-----END PRIVATE KEY-----\n",
    "client_email" : "example-service-account@example-project-123456.iam.gserviceaccount.com"
  })
  registry = "gcr.io"

  workspace_id = var.seqera_workspace_id
}

# Google Artifact Registry (GAR) credential
resource "seqera_container_registry_credential" "example_gar" {
  name      = "example-gar-credentials"
  user_name = "_json_key"
  password = jsonencode({
    "type" : "service_account",
    "project_id" : "example-project-123456",
    "private_key_id" : "1234567890abcdef1234567890abcdef12345678",
    "private_key" : "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7xYz2LqYqrLYS\nexamplekeycontenthere/notarealkey\n-----END PRIVATE KEY-----\n",
    "client_email" : "example-service-account@example-project-123456.iam.gserviceaccount.com"
  })
  registry = "us-east1-docker.pkg.dev"

  workspace_id = var.seqera_workspace_id
}

# Azure Container Registry (ACR) credential
resource "seqera_container_registry_credential" "example_acr" {
  name      = "example-acr-credentials"
  user_name = "example-acr-username"
  password  = "example-acr-password-123456"
  registry  = "exampleregistry.azurecr.io"

  workspace_id = var.seqera_workspace_id
}

# GitHub Container Registry (GHCR) credential
resource "seqera_container_registry_credential" "example_ghcr" {
  name      = "example-ghcr-credentials"
  user_name = "example-github-user"
  password  = "ghp_ExamplePersonalAccessToken123456789ABCDEFGHIJ"
  registry  = "ghcr.io"

  workspace_id = var.seqera_workspace_id
}

# GitLab Container Registry credential
resource "seqera_container_registry_credential" "example_gitlab_registry" {
  name      = "example-gitlab-registry-credentials"
  user_name = "example-gitlab-user"
  password  = "glpat-ExamplePersonalAccessToken1234567890AB"
  registry  = "registry.gitlab.com"

  workspace_id = var.seqera_workspace_id
}

# Quay.io credential
resource "seqera_container_registry_credential" "example_quay" {
  name      = "example-quay-io-credentials"
  user_name = "example-quay-user"
  password  = "example-quay-password-123456"
  registry  = "quay.io"

  workspace_id = var.seqera_workspace_id
}

# Private/Custom Container Registry credential
resource "seqera_container_registry_credential" "example_private" {
  name      = "example-private-registry-credentials"
  user_name = "example-private-user"
  password  = "example-private-password-123456"
  registry  = "registry.example.com"

  workspace_id = var.seqera_workspace_id
}

# Output the credential ID
output "container_registry_credential_id" {
  value       = seqera_container_registry_credential.example_docker_hub.credentials_id
  description = "The ID of the Container Registry credential"
}

output "container_registry_credential_provider_type" {
  value       = seqera_container_registry_credential.example_docker_hub.provider_type
  description = "The provider type (should be 'container-registry')"
}
