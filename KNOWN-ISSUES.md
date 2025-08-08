  # Known Issues

  This document outlines known issues, limitations, and workarounds for the Seqera Terraform Provider. The provider is auto-generated using Speakeasy from OpenAPI
  specifications, which can sometimes result in specific behaviors that users should be aware of.

## Reporting and Tracking Issues

For additional known issues and bug reports, please check the [GitHub Issues](https://github.com/seqeralabs/terraform-provider-seqera/issues) page. Users should search through existing GitHub issues as they may contain more up-to-date information about current problems, workarounds, and status updates.

## Import Limitations

### Import Functionality Work in Progress
Import functionality for most resources is currently work in progress. The following resources do not yet support importing existing infrastructure into Terraform state:

- `seqera_action`
- `seqera_compute_env`
- `seqera_credential`
- `seqera_data_link`
- `seqera_datasets`
- `seqera_labels`
- `seqera_pipeline`
- `seqera_pipeline_secret`
- `seqera_studios`
- `seqera_tokens`
- `seqera_workflows`

This functionality is being actively developed and will be available in future releases.
