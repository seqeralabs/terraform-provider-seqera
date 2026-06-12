# Seqera Platform Terraform Provider

> [!NOTE] > **Public Preview** - This provider is currently in public preview.
>
> We'd love your feedback! Please test the provider with your use cases and [report any issues](https://github.com/seqeralabs/terraform-provider-seqera/issues) you encounter. Your input will help us build a better stable release.

> [!IMPORTANT] > **Built for Seqera Organizations** — This provider targets admins managing shared resources within a Seqera organization. Personal-workspace use isn't part of our supported scope and some resources will not work in that context.

> [!CAUTION] > **Deprecated Resources** - Resources marked as deprecated should be avoided in new configurations, as they will be removed in the future release. Please migrate to their recommended replacements.

Terraform Provider for the Seqera Platform API.

<div align="left">
    <a href="https://www.speakeasy.com/?utm_source=seqera&utm_campaign=terraform"><img src="https://custom-icon-badges.demolab.com/badge/-Built%20By%20Speakeasy-212015?style=for-the-badge&logoColor=FBE331&logo=speakeasy&labelColor=545454" /></a>
    <a href="https://opensource.org/licenses/MIT">
        <img src="https://img.shields.io/badge/License-MIT-blue.svg" style="width: 100px; height: 28px;" />
    </a>
</div>

<!-- Start Summary [summary] -->
## Summary

Seqera API: The Seqera Platform Terraform Provider enables infrastructure-as-code management of Seqera Platform resources. This provider allows you to programmatically create, configure, and manage organizations, workspaces, compute environments, pipelines, credentials, and other Seqera Platform components using Terraform.
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [Seqera Platform Terraform Provider](#seqera-platform-terraform-provider)
  * [Installation](#installation)
  * [Authentication](#authentication)
  * [Available Resources and Data Sources](#available-resources-and-data-sources)
  * [Examples](#examples)
  * [Testing the provider locally](#testing-the-provider-locally)
* [Development](#development)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start Installation [installation] -->
## Installation

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    seqera = {
      source  = "seqeralabs/seqera"
      version = "0.40.0"
    }
  }
}

provider "seqera" {
  server_url = "..." # Optional
}
```
<!-- End Installation [installation] -->

<!-- Start Authentication [security] -->
## Authentication

This provider supports authentication configuration via environment variables and provider configuration.

The configuration precedence is:

- Provider configuration
- Environment variables

Available configuration:

| Provider Attribute | Description |
|---|---|
| `bearer_auth` | HTTP Bearer. Configurable via environment variable `TOWER_ACCESS_TOKEN`. |
<!-- End Authentication [security] -->

<!-- Start Available Resources and Data Sources [operations] -->
## Available Resources and Data Sources

### Managed Resources

* [seqera_aws_batch_ce](docs/resources/aws_batch_ce.md)
* [seqera_aws_compute_env](docs/resources/aws_compute_env.md)
* [seqera_aws_credential](docs/resources/aws_credential.md)
* [seqera_action](docs/resources/action.md)
* [seqera_aws_cloud_ce](docs/resources/aws_cloud_ce.md)
* [seqera_azure_batch_ce](docs/resources/azure_batch_ce.md)
* [seqera_azure_cloud_ce](docs/resources/azure_cloud_ce.md)
* [seqera_azure_cloud_credential](docs/resources/azure_cloud_credential.md)
* [seqera_azure_credential](docs/resources/azure_credential.md)
* [seqera_azure_entra_credential](docs/resources/azure_entra_credential.md)
* [seqera_bitbucket_credential](docs/resources/bitbucket_credential.md)
* [seqera_codecommit_credential](docs/resources/codecommit_credential.md)
* [seqera_compute_env](docs/resources/compute_env.md)
* [seqera_container_registry_credential](docs/resources/container_registry_credential.md)
* [seqera_credential](docs/resources/credential.md)
* [seqera_custom_role](docs/resources/custom_role.md)
* [seqera_data_link](docs/resources/data_link.md)
* [seqera_datasets](docs/resources/datasets.md)
* [seqera_gcp_batch_ce](docs/resources/gcp_batch_ce.md)
* [seqera_gcp_cloud_ce](docs/resources/gcp_cloud_ce.md)
* [seqera_gitea_credential](docs/resources/gitea_credential.md)
* [seqera_github_credential](docs/resources/github_credential.md)
* [seqera_gitlab_credential](docs/resources/gitlab_credential.md)
* [seqera_google_credential](docs/resources/google_credential.md)
* [seqera_kubernetes_credential](docs/resources/kubernetes_credential.md)
* [seqera_labels](docs/resources/labels.md)
* [seqera_managed_compute_ce](docs/resources/managed_compute_ce.md)
* [seqera_orgs](docs/resources/orgs.md)
* [seqera_pipeline](docs/resources/pipeline.md)
* [seqera_pipeline_secret](docs/resources/pipeline_secret.md)
* [seqera_primary_compute_env](docs/resources/primary_compute_env.md)
* [seqera_ssh_credential](docs/resources/ssh_credential.md)
* [seqera_studios](docs/resources/studios.md)
* [seqera_teams](docs/resources/teams.md)
* [seqera_tokens](docs/resources/tokens.md)
* [seqera_tower_agent_credential](docs/resources/tower_agent_credential.md)
* [seqera_workflows](docs/resources/workflows.md)
* [seqera_workspace](docs/resources/workspace.md)

### Data Sources

* [seqera_credentials](docs/data-sources/credentials.md)
* [seqera_data_links](docs/data-sources/data_links.md)
<!-- End Available Resources and Data Sources [operations] -->

<!-- Start Examples [examples] -->

## Examples

The `examples/terraform-examples` directory contains comprehensive Terraform configurations demonstrating how to use the Seqera Platform provider across different cloud platforms. Each example includes a complete setup from organization to running nf-core/rnaseq.

### Cloud Platform Examples

- **[AWS Example (`examples/terraform-examples/aws/`)](examples/terraform-examples/aws/README.md)** - Complete AWS Batch setup with nf-core/rnaseq pipeline
- **[Azure Example (`examples/terraform-examples/azure/`)](examples/terraform-examples/azure/README.md)** - Complete Azure Batch setup with nf-core/rnaseq pipeline
- **[GCP Example (`examples/terraform-examples/gcp/`)](examples/terraform-examples/gcp/README.md)** - Complete Google Batch setup with genomics-optimized instances

### Getting Started with Examples

1. **Choose your cloud platform** from `examples/terraform-examples/aws/`, `examples/terraform-examples/azure/`, or `examples/terraform-examples/gcp/`
2. **Copy the example tfvars**: `cp terraform.tfvars.example terraform.tfvars`
3. **Configure your credentials** and settings in `terraform.tfvars`
4. **Amend any variable/resource names or values** ,ensure you update your organization name as that has to be unique.
5. **Initialize Terraform**: `terraform init`
6. **Review the plan**: `terraform plan`
7. **Apply when ready**: `terraform apply`

Each example includes detailed variable descriptions and validation rules to help you configure the resources correctly for your environment.

<!-- End Examples [examples] -->

<!-- Start Testing the provider locally [usage] -->
## Testing the provider locally

#### Local Provider

Should you want to validate a change locally, the `--debug` flag allows you to execute the provider against a terraform instance locally.

This also allows for debuggers (e.g. delve) to be attached to the provider.

```sh
go run main.go --debug
# Copy the TF_REATTACH_PROVIDERS env var
# In a new terminal
cd examples/your-example
TF_REATTACH_PROVIDERS=... terraform init
TF_REATTACH_PROVIDERS=... terraform apply
```

#### Compiled Provider

Terraform allows you to use local provider builds by setting a `dev_overrides` block in a configuration file called `.terraformrc`. This block overrides all other configured installation methods.

1. Execute `go build` to construct a binary called `terraform-provider-seqera`
2. Ensure that the `.terraformrc` file is configured with a `dev_overrides` section such that your local copy of terraform can see the provider binary

Terraform searches for the `.terraformrc` file in your home directory and applies any configuration settings you set.

```
provider_installation {

  dev_overrides {
      "registry.terraform.io/seqeralabs/seqera" = "<PATH>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```
<!-- End Testing the provider locally [usage] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Contributions

While we value open-source contributions to this terraform provider, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation.
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release.

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=seqera&utm_campaign=terraform)
