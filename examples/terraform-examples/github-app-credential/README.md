# GitHub App credential

This configuration creates a GitHub App credential in an existing Seqera
workspace. It does not create or configure the GitHub App itself.

## Prerequisites

- Terraform 1.11 or later
- A Seqera access token with permission to manage credentials in the target
  workspace
- An installed GitHub App, with its App ID, Client ID, and PEM private key

## Run it

```shell
cd examples/terraform-examples/github-app-credential
cp terraform.tfvars.example terraform.tfvars
export TOWER_ACCESS_TOKEN='...'
terraform init
terraform plan
terraform apply
```

Set `workspace_id`, `github_app_id`, `github_app_client_id`, and
`github_app_private_key_path` in `terraform.tfvars`. The private key stays in
the local PEM file and is not written to `terraform.tfvars`.

Set `github_base_url` to the organization or repository URL on which the App
is installed.
