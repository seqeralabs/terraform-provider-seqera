  # Known Issues

  This document outlines known issues, limitations, and workarounds for the Seqera Terraform Provider. The provider is auto-generated using Speakeasy from OpenAPI
  specifications, which can sometimes result in specific behaviors that users should be aware of.

## Reporting and Tracking Issues

For additional known issues and bug reports, please check the [GitHub Issues](https://github.com/seqeralabs/terraform-provider-seqera/issues) page. Users should search through existing GitHub issues as they may contain more up-to-date information about current problems, workarounds, and status updates.

## Import Limitations

### Import Functionality Work in Progress
Import functionality for most resources is currently work in progress. The following resources do not yet support importing existing infrastructure into Terraform state:

- `seqera_action`
- `seqera_credential`
- `seqera_data_link`
- `seqera_datasets`
- `seqera_labels`
- `seqera_pipeline_secret`
- `seqera_studios`
- `seqera_tokens`
- `seqera_workflows`

This functionality is being actively developed and will be available in future releases.

**Note**: Some resources that support import may require workspace context in JSON format (e.g., `'{"resource_id": "abc", "workspace_id": 123}'`). Check the resource documentation for the exact import syntax.

## API Permission Issues

### 403 Forbidden Errors
Users may occasionally receive 403 Forbidden errors when working with this provider:

```
Error: unexpected response from API. Got an unexpected response code 403

**Request**:
GET /pipelines/17409243925855 HTTP/1.1
Host: api.cloud.seqera.io
Accept: application/json
Authorization: (sensitive)
User-Agent: speakeasy-sdk/terraform 0.0.3 2.675.0 1.56.0 github.com/speakeasy/terraform-provider-seqera/internal/sdk

**Response**:
HTTP/2.0 403 Forbidden
Date: Fri, 08 Aug 2025 14:44:33 GMT
Content-Length: 0
```

If you encounter 403 errors, please create a bug report with your example configuration and the complete error message.

### Other API Errors
Users may encounter various API errors with detailed response information:

```
Error: unexpected response from API. Got an unexpected response code 409

  with seqera_orgs.test_org,
  on workspace.tf line 38, in resource "seqera_orgs" "test_org":
  38: resource "seqera_orgs" "test_org" {

**Request**:
POST /orgs HTTP/1.1
Host: api.cloud.seqera.io
Accept: application/json
Authorization: (sensitive)
Content-Type: application/json
User-Agent: speakeasy-sdk/terraform 0.0.3 2.634.2 1.45.0 github.com/speakeasy/terraform-provider-seqera/internal/sdk

**Response**:
HTTP/2.0 409 Conflict
Content-Length: 90
Content-Type: application/json
Date: Tue, 24 Jun 2025 19:22:16 GMT

{"message":"An organization with name 'seqera_test_shahbaz_tf_provider_1' already exists"}
```

If you encounter API errors that are not documented as known issues, please create a bug report including the complete error message and your Terraform configuration.

## Resource State Issues

### Credentials Drift Detection
The `seqera_credential` resource may show continuous updates in Terraform plans due to state drift detection issues:

```hcl
# module.aws_batch.seqera_credential.aws_credential will be updated in-place
~ resource "seqera_credential" "aws_credential" {
    ~ keys           = {
        ~ aws = {
            + secret_key      = (sensitive value)
              # (2 unchanged attributes hidden)
          }
      }
      name           = "test_credential"
      # (7 unchanged attributes hidden)
  }
```

This is a known issue and work in progress. The resource will continue to function correctly despite these continuous updates.
