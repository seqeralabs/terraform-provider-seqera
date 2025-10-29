# Seqera Credentials Test Examples

This directory contains example Terraform configurations for all supported credential types in the Seqera Terraform Provider. These examples use non-sensitive placeholder values for testing purposes.

## Overview

The Seqera Terraform Provider supports the following credential types:

| Credential Type | File | Provider Type |
|----------------|------|---------------|
| AWS | `aws_credential.tf` | `aws` |
| Azure | `azure_credential.tf` | `azure` |
| Google Cloud | `google_credential.tf` | `google` |
| SSH | `ssh_credential.tf` | `ssh` |
| Kubernetes | `kubernetes_credential.tf` | `k8s` |
| GitHub | `github_credential.tf` | `github` |
| GitLab | `gitlab_credential.tf` | `gitlab` |
| Gitea | `gitea_credential.tf` | `gitea` |
| Bitbucket | `bitbucket_credential.tf` | `bitbucket` |
| CodeCommit | `codecommit_credential.tf` | `codecommit` |
| Container Registry | `container_registry_credential.tf` | `container-registry` |
| Tower Agent | `tower_agent_credential.tf` | `tower-agent` |

## Prerequisites

1. **Seqera Platform Access**: You need access to a Seqera Platform instance (Cloud or Enterprise)
2. **Access Token**: Generate an access token from your Seqera Platform account
3. **Terraform**: Install Terraform v1.0 or later

## Configuration

### Environment Variables

The easiest way to configure the provider is using environment variables:

```bash
export SEQERA_API_URL="https://api.cloud.seqera.io"
export SEQERA_ACCESS_TOKEN="your-access-token-here"
```

### Provider Configuration

Alternatively, configure the provider directly in `main.tf`:

```hcl
provider "seqera" {
  api_url      = "https://api.cloud.seqera.io"
  access_token = "your-access-token-here"
}
```

## Usage

### Initialize Terraform

```bash
terraform init
```

### Validate Configuration

```bash
terraform validate
```

### Plan Changes

To see what resources will be created without actually creating them:

```bash
terraform plan
```

### Apply Configuration

⚠️ **Warning**: The example credentials contain placeholder values. These will create resources in your Seqera Platform but may not be functional for actual workloads.

To create all credential resources:

```bash
terraform apply
```

To create specific credential resources:

```bash
# Apply only AWS credentials
terraform apply -target=seqera_aws_credential.example_basic

# Apply only Azure credentials
terraform apply -target=seqera_azure_credential.example_shared_key

# Apply multiple specific resources
terraform apply \
  -target=seqera_aws_credential.example_basic \
  -target=seqera_google_credential.example
```

### Destroy Resources

To remove all created credentials:

```bash
terraform destroy
```

To destroy specific resources:

```bash
terraform destroy -target=seqera_aws_credential.example_basic
```

## Testing Individual Credential Types

If you want to test only specific credential types:

1. **Comment out other resources**: Edit the `.tf` files and comment out resources you don't want to create
2. **Use targeted apply**: Use the `-target` flag as shown above
3. **Create a separate test directory**: Copy only the credential files you want to test

### Example: Test Only Cloud Provider Credentials

```bash
terraform apply \
  -target=seqera_aws_credential.example_basic \
  -target=seqera_azure_credential.example_shared_key \
  -target=seqera_google_credential.example
```

## Workspace Association

All credential examples include commented-out workspace association. To associate credentials with a specific workspace:

1. Uncomment the workspace data source in `main.tf`:
   ```hcl
   data "seqera_workspace" "example" {
     name   = "your-workspace-name"
     org_id = 1
   }
   ```

2. Uncomment the `workspace_id` parameter in each credential resource:
   ```hcl
   resource "seqera_aws_credential" "example_basic" {
     name         = "Example AWS Credentials"
     access_key   = "AKIAIOSFODNN7EXAMPLE"
     secret_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
     workspace_id = data.seqera_workspace.example.id  # Uncomment this line
   }
   ```

## Using Real Credentials

⚠️ **Security Warning**: Never commit real credentials to version control!

To use real credentials, consider these approaches:

### 1. Environment Variables

```hcl
resource "seqera_aws_credential" "example" {
  name       = "Production AWS Credentials"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}
```

### 2. Terraform Variables

Create a `terraform.tfvars` file (add to `.gitignore`):

```hcl
aws_access_key = "AKIA..."
aws_secret_key = "..."
```

### 3. External Secret Management

Use tools like:
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Google Secret Manager

Example with AWS Secrets Manager:

```hcl
data "aws_secretsmanager_secret_version" "aws_creds" {
  secret_id = "seqera/aws-credentials"
}

locals {
  aws_creds = jsondecode(data.aws_secretsmanager_secret_version.aws_creds.secret_string)
}

resource "seqera_aws_credential" "production" {
  name       = "Production AWS Credentials"
  access_key = local.aws_creds.access_key
  secret_key = local.aws_creds.secret_key
}
```

## Outputs

Each credential example includes outputs for:
- `{type}_credential_id`: The unique identifier for the credential
- `{type}_credential_provider_type`: The provider type (for verification)

View outputs after applying:

```bash
terraform output
```

View specific output:

```bash
terraform output aws_credential_id
```

## Data Sources

Each credential type includes a corresponding data source example for reading existing credentials:

```hcl
data "seqera_aws_credential" "example" {
  id = seqera_aws_credential.example_basic.credentials_id
}
```

## Important Notes

1. **Non-Functional Credentials**: The example credentials are placeholders and will not work for actual compute operations
2. **Validation**: Some credential types may fail validation if Seqera Platform attempts to verify them
3. **Cleanup**: Remember to destroy test resources to avoid clutter in your Seqera Platform instance
4. **Rate Limits**: Be aware of API rate limits when creating many credentials at once
5. **Workspace Permissions**: Ensure you have appropriate permissions to create credentials in the target workspace

## Troubleshooting

### Authentication Errors

If you see authentication errors:
- Verify your `SEQERA_ACCESS_TOKEN` is valid
- Check the `SEQERA_API_URL` matches your Seqera Platform instance
- Ensure the token has appropriate permissions

### Resource Creation Failures

If resource creation fails:
- Check the Terraform plan output for errors
- Verify the credential format matches the expected schema
- Review Seqera Platform logs for validation errors

### Provider Not Found

If Terraform can't find the provider:
```bash
terraform init -upgrade
```

## Additional Resources

- [Seqera Terraform Provider Documentation](https://registry.terraform.io/providers/seqeralabs/seqera/latest/docs)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Terraform Documentation](https://www.terraform.io/docs)

## Example Workflow

A typical testing workflow:

```bash
# 1. Set up environment
export SEQERA_API_URL="https://api.cloud.seqera.io"
export SEQERA_ACCESS_TOKEN="your-token"

# 2. Initialize Terraform
terraform init

# 3. Validate configuration
terraform validate

# 4. Preview changes
terraform plan

# 5. Create specific credential for testing
terraform apply -target=seqera_aws_credential.example_basic

# 6. Verify in Seqera Platform UI
# Navigate to Credentials section in your workspace

# 7. Clean up when done
terraform destroy
```

## Contributing

When adding new credential examples:
1. Use non-sensitive placeholder values
2. Include both basic and advanced examples where applicable
3. Add data source examples
4. Include outputs for credential ID and provider type
5. Document any special considerations in this README
