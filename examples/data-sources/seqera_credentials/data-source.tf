# Basic usage - fetch all credentials in a workspace
data "seqera_credentials" "all" {
  workspace_id = seqera_workspace.my_workspace.id
}

# Create a map indexed by credential name for easy lookup
locals {
  credentials = {
    for cred in data.seqera_credentials.all.credentials : cred.name => cred
  }
}

# Access a specific credential by name
output "aws_prod_credential_id" {
  value = local.credentials["production-aws-account"].id
}

# Filter credentials by provider type
locals {
  aws_credentials = {
    for cred in data.seqera_credentials.all.credentials : cred.name => cred
    if cred.provider == "aws"
  }

  github_credentials = {
    for cred in data.seqera_credentials.all.credentials : cred.name => cred
    if cred.provider == "github"
  }
}

# Use filtered credentials in resources
resource "seqera_aws_batch_ce" "example" {
  name           = "my-compute-env"
  workspace_id   = seqera_workspace.my_workspace.id
  credentials_id = local.credentials["production-aws-account"].id
  # ... other configuration
}
