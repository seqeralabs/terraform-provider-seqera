# Known Issues

This document outlines known issues, limitations, and workarounds for the Seqera Terraform Provider. The provider is auto-generated using Speakeasy from OpenAPI
specifications, which can sometimes result in specific behaviors that users should be aware of.

## Reporting and Tracking Issues

For additional known issues and bug reports, please check the [GitHub Issues](https://github.com/seqeralabs/terraform-provider-seqera/issues) page. Users should search through existing GitHub issues as they may contain more up-to-date information about current problems, workarounds, and status updates.

## Import Limitations

### Import Functionality Work in Progress

Import functionality for the following resources is not yet implemented:

- `seqera_datasets`
- `seqera_labels`
- `seqera_tokens`

This functionality is being actively developed and will be available in future releases.

**Note**: Some resources that support import may require workspace context in JSON format (e.g., `'{"resource_id": "abc", "workspace_id": 123}'`). Check the resource documentation for the exact import syntax.
