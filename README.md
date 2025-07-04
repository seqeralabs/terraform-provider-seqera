# Seqera Platform Terraform Provider

Terraform Provider for the Seqera Platform API.

<div align="left">
    <a href="https://www.speakeasy.com/?utm_source=seqera&utm_campaign=terraform"><img src="https://custom-icon-badges.demolab.com/badge/-Built%20By%20Speakeasy-212015?style=for-the-badge&logoColor=FBE331&logo=speakeasy&labelColor=545454" /></a>
    <a href="https://opensource.org/licenses/MIT">
        <img src="https://img.shields.io/badge/License-MIT-blue.svg" style="width: 100px; height: 28px;" />
    </a>
</div>


<!-- Start Summary [summary] -->
## Summary

Th Seqera terraform provider enables Terraform to manage [Seqera platform](https://cloud.seqera.io) resources. For more information on Seqera Platform please check out our [documentation here](https://docs.seqera.io/). 

<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [Installation](#installation)
* [Available Resources and Data Sources](#available-resources-and-data-sources)
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
      source  = "speakeasy/seqera"
      version = "0.0.3"
    }
  }
}

provider "seqera" {
  # Configuration options
}
```
<!-- End Installation [installation] -->

<!-- Start Available Resources and Data Sources [operations] -->
## Available Resources and Data Sources

### Resources

For an updated list of available resources and relevant documentation please go to the [documetation](docs/resources/)
### Data Sources

For an updated list of available resources and relevant documentation please go to the [documetation](docs/data-sources/)

<!-- End Available Resources and Data Sources [operations] -->

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
      "registry.terraform.io/speakeasy/seqera" = "<PATH>"
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
