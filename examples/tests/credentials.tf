resource "seqera_credential" "gcp_credential" {
  name          = "test-gcp-credential"
  provider_type = "google"
  description   = "Google Cloud credentials for test project"
  workspace_id  = local.workspace_id
  
  keys = {
    google = {
      data          = jsonencode({
        type                        = "service_account"
        project_id                 = "test-project-123456"
        private_key_id            = "key-id-12345"
        private_key               = "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC..."
        client_email              = "test-service@test-project-123456.iam.gserviceaccount.com"
        client_id                 = "123456789012345678901"
        auth_uri                  = "https://accounts.google.com/o/oauth2/auth"
        token_uri                 = "https://oauth2.googleapis.com/token"
        auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
        client_x509_cert_url      = "https://www.googleapis.com/robot/v1/metadata/x509/test-service%40test-project-123456.iam.gserviceaccount.com"
      })
      discriminator = "google"
    }
  }
}


resource "seqera_credential" "azure_credential" {
  name          = "test-azure-credential"
  provider_type = "azure"
  description   = "Azure Batch and Storage credentials"
  workspace_id  = local.workspace_id
  
  keys = {
    azure = {
      batch_name    = "testbatchaccount"
      batch_key     = "batch-account-key-example"
      storage_name  = "teststorageaccount"
      storage_key   = "storage-account-key-example"
      discriminator = "azure"
    }
  }
}

resource "seqera_credential" "github_credential" {
  name          = "test-github-credential"
  provider_type = "github"
  description   = "GitHub access credentials"
  workspace_id  = local.workspace_id
  base_url      = "https://github.com"
  
  keys = {
    github = {
      username      = "test-user-github"
      password      = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
      discriminator = "github"
    }
  }
}

resource "seqera_credential" "ssh_credential" {
  name          = "test-ssh-credential"
  provider_type = "ssh"
  description   = "SSH key for remote access"
  workspace_id  = local.workspace_id
  
  keys = {
    ssh = {
      private_key   = file("~/.ssh/id_rsa")
      passphrase    = "my-secure-passphrase"
      discriminator = "ssh"
    }
  }
}

# resource "seqera_credential" "k8s_credential" {
#   name          = "test-k8s-credential"
#   provider_type = "k8s"
#   description   = "Kubernetes cluster access"
#   workspace_id  = local.workspace_id
  
#   keys = {
#     k8s = {
#       certificate   = base64encode(file("~/.kube/ca.crt"))
#       private_key   = base64encode(file("~/.kube/client.key"))
#       token         = "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9..."
#       discriminator = "k8s"
#     }
#   }
# }

## Need valid credentials otherwise this returns an invalid credentials error on creation
# │ Error: unexpected response from API. Got an unexpected response code 400
# │
# │   with seqera_credential.container_registry_credential,
# │   on credentials.tf line 134, in resource "seqera_credential" "container_registry_credential":
# │  134: resource "seqera_credential" "container_registry_credential" {
# │
# │ **Request**:
# │ POST /credentials?workspaceId=49242724423913 HTTP/1.1
# │ Host: api.cloud.seqera.io
# │ Accept: application/json
# │ Authorization: (sensitive)
# │ Content-Type: application/json
# │ User-Agent: speakeasy-sdk/terraform 0.0.3 2.632.2 1.45.0 github.com/speakeasy/terraform-provider-seqera/internal/sdk
# │
# │
# │ **Response**:
# │ HTTP/2.0 400 Bad Request
# │ Content-Length: 68
# │ Content-Type: application/json
# │ Date: Mon, 23 Jun 2025 13:24:53 GMT
# │
# │ {"message":"Invalid credentials for container registry 'docker.io'"}

# resource "seqera_credential" "container_registry_credential" {
#   name          = "test-docker-credential"
#   provider_type = "container-reg"
#   description   = "Docker Hub registry access"
#   workspace_id  = local.workspace_id
  
#   keys = {
#     container_reg = {
#       registry      = "docker.io"
#       user_name     = "test-docker-user"
#       password      = "docker-hub-access-token"
#       discriminator = "container-reg"
#     }
#   }
# }

resource "seqera_credential" "gitlab_credential" {
  name          = "test-gitlab-credential"
  provider_type = "gitlab"
  description   = "GitLab repository access"
  workspace_id  = local.workspace_id
  base_url      = "https://gitlab.com"
  
  keys = {
    gitlab = {
      username      = "test-user-gitlab"
      password      = "personal-access-token-example"
      token         = "glpat-xxxxxxxxxxxxxxxxxxxx"
      discriminator = "gitlab"
    }
  }
}