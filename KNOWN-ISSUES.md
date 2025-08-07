  # Known Issues

  This document outlines known issues, limitations, and workarounds for the Seqera Terraform Provider. The provider is auto-generated using Speakeasy from OpenAPI
  specifications, which can sometimes result in specific behaviors that users should be aware of.

## Reporting and Tracking Issues

For additional known issues and bug reports, please check the [GitHub Issues](https://github.com/seqeralabs/terraform-provider-seqera/issues) page. Users should search through existing GitHub issues as they may contain more up-to-date information about current problems, workarounds, and status updates.

## Import Limitations

### Tokens and Labels Resources
It is currently not possible to import existing `seqera_tokens` and `seqera_labels` resources into Terraform state. These resources lack the necessary individual GET endpoints in the Seqera Platform API specification, which are required for the import functionality to work properly.
