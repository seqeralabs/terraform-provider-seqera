# v0.26.6

FIX

- **Cloud / 25.3**

- **Compute Resources** - Support the new deleting lifecycle phase for Compute environments to avoid plan / state errors.

# v0.26.5

FIX:

- **Credentials Resources** - Fixed an issue where the `base_url` field was not being returned in API responses for GitHub, GitLab, Gitea, Bitbucket, and CodeCommit credentials, preventing the URL from displaying correctly in the Seqera Platform UI.
- **GitHub Credentials** - Fixed an issue where the GitHub Personal Access Token field was using incorrect API field name `accessToken` instead of `password`, resulting in invalid credentials.
- **CodeCommit Credentials** - Fixed an issue where AWS credential fields were using incorrect API field names `accessKey`/`secretKey` instead of `username`/`password`, resulting in authentication failures.
- **Container Registry Credentials** - Fixed an issue where the `registry` field was incorrectly marked as write-only, preventing the registry URL from being readable in API responses.
- **Google Cloud Credentials** - Fixed critical issue where the service account JSON (`data` field) was not being sent in API requests, causing credential creation to fail. Added internal `keyType` field to SDK models to enable proper code generation while keeping it hidden from Terraform schema and documentation.
- **Kubernetes Credentials** - Fixed critical issue where authentication fields (`token`, `certificate`, `private_key`) were not being sent in API requests, causing credential creation to fail. Added internal `keyType` field to SDK models to enable proper code generation while keeping it hidden from Terraform schema and documentation.
- **SSH Credentials** - Improved implementation by hiding internal `key_type` field from Terraform schema and documentation while maintaining correct API request generation. This field is now only present in SDK models for code generation purposes.

# v0.26.4

FIX:

- **Compute Environments** Added validation for compute and head job targetting of environment variables.
- **AWS Credentials** Allowed the ommission of Secret Key & Access Key values when using a role.

# v0.26.3

FEATURES:

- **Seqera Action Resource** Cleaned up the resource removing unused fields.

FIX:

- **Seqera Credentials Resource** Added missing username fields.

# v0.26.2

FEATURES:

- **New Data Source:** `seqera_credentials` - Lists all credentials with optional workspace filtering. Returns credential `id`, `name`, and `provider_type` for each credential. Use Terraform locals with `for` expressions to filter by provider type or name (e.g., `local.creds["credential-name"].id`)
- **New Data Source:** `seqera_data_links` - Lists all data links with optional workspace filtering. Returns data link `id`, `name`, `provider`, `resource_ref`, and `region` for each data link. Use Terraform locals with `for` expressions to filter by provider type, region, or name:

  ```hcl
  data "seqera_data_links" "all" {
    workspace_id = seqera_workspace.my_workspace.id
  }

  locals {
    # Index by name for easy lookup
    datalinks = {
      for dl in data.seqera_data_links.all.data_links : dl.name => dl
    }

    # Filter AWS data links in us-east-1
    aws_us_east_1 = {
      for dl in data.seqera_data_links.all.data_links : dl.name => dl
      if dl.provider == "aws" && dl.region == "us-east-1"
    }

    # Filter by provider
    aws_datalinks = {
      for dl in data.seqera_data_links.all.data_links : dl.name => dl
      if dl.provider == "aws"
    }
  }

  # Access: local.datalinks["my-s3-bucket"].id
  ```

ENHANCEMENTS:

- **Data Sources**: Removed automatic data source generation for all resources. Resources now only support the read operation for state management. This simplifies the provider API surface and reduces confusion between resources and data sources.

- **AWS Batch Compute Environments**: Updated `dispose_on_deletion` documentation to clarify that AWS credentials must have appropriate permissions to delete resources (Batch compute environments, job queues, launch templates, IAM roles, instance profiles, FSx/EFS file systems) when this flag is enabled.

# v0.26.1

FEATURES:

- **Credentials**: Credentials now use `.id` as an identifier vs `.credentials_id` you will have to update references to these in the code base and use terraform refresh.

- **Compute Environments**: Credentials now use `.id` as an identifier vs `.compute_env_id` you will have to update references to these in the code base and use terraform refresh.

ENHANCEMENTS:

- **Studios**: The `configuration` block is now required to prevent backend errors. GPU field defaults to 0 (disabled) when not specified.

- **Studios**: Added `environment` field in configuration for setting studio-specific environment variables. Variable names must contain only alphanumeric and underscore characters, and cannot begin with a number.

- **Studios**: Added varios examples showing:

  - Minimal studio with empty configuration
  - Conda environment setup using both heredoc and yamlencode() approaches
  - Resource label integration
  - Mounted data configuration
  - Custom environment variables

- **Studios**: GPU field now has clear description: "Set to 0 to disable GPU or 1 to enable GPU"

# v0.26.0

FEATURES:

- **New Resource:** `seqera_aws_batch_ce` - AWS Batch-specific compute environment resource
- **New Resource:** `seqera_aws_credential` - AWS credentials
- **New Resource:** `seqera_azure_credential` - Azure credentials
- **New Resource:** `seqera_bitbucket_credential` - Bitbucket credentials
- **New Resource:** `seqera_codecommit_credential` - AWS CodeCommit credentials
- **New Resource:** `seqera_container_registry_credential` - Container registry credentials
- **New Resource:** `seqera_gitea_credential` - Gitea credentials
- **New Resource:** `seqera_github_credential` - GitHub credentials
- **New Resource:** `seqera_gitlab_credential` - GitLab credentials
- **New Resource:** `seqera_google_credential` - Google Cloud Platform credentials
- **New Resource:** `seqera_kubernetes_credential` - Kubernetes credentials
- **New Resource:** `seqera_ssh_credential` - SSH credentials
- **New Resource:** `seqera_tower_agent_credential` - Tower Agent credentials

ENHANCEMENTS:

- **Wave validation**: When `enable_wave` is set to `true`, `enable_fusion` must be explicitly configured (cannot be null). Wave containers work with or without Fusion2, but the configuration must be explicit to avoid ambiguity.

- **Fusion validation**: Enforces two key rules for AWS Batch configurations:
  - When Fusion2 (`enable_fusion=true`) is enabled, Wave (`enable_wave=true`) must also be enabled, as Fusion2 depends on Wave for container management
  - When both Forge and Fusion2 are enabled, `cli_path` must not be set, as Forge manages the CLI path automatically
- Compute environment behaviour mirrors platform UI

- **Label name validation**: Label names must be 1-39 alphanumeric characters, can contain dashes (`-`) or underscores (`_`) as separators, and must start and end with alphanumeric characters (e.g., `environment`, `my-label`, `test_123`)

- **Label default validation**: The `is_default` attribute can only be set to `true` when `resource` is also `true`, as only resource labels can be automatically applied to new resources

- **Schema cleanup for `seqera_pipeline`**: Removed 20+ runtime and computed fields that should not be managed by Terraform:

  - Removed transient fields: `userLastName`, `orgId`, `orgName`, `workspaceName`, `deleted`, `lastUpdated`, `labels`, `computeEnv`, optimization-related fields
  - Removed computed fields: `visibility` (inherited from workspace), repository metadata fields (discovered from git repository)
  - Cleaned up `launch` block to only include user-configurable fields from Seqera Platform UI

- **Schema cleanup for `seqera_studios`**: Removed 20+ runtime and transient fields that should not be managed by Terraform:

  - Removed runtime state: `user`, `studioUrl`, `computeEnv`, `template`, `statusInfo`, `activeConnections`, `progress`
  - Removed timestamps: `dateCreated`, `lastUpdated`, `lastStarted`
  - Removed computed fields: `effectiveLifespanHours`, `waveBuildUrl`, `baseImage`, `customImage`, `mountedDataLinks`, `labels`
  - Removed checkpoint references: `parentCheckpoint`

- **Schema cleanup for `seqera_workflows`**: Removed 30+ runtime and execution fields that should not be managed by Terraform:

  - Removed runtime execution data: `progress`, `messages`, `jobInfo`, `platform`, `optimized`
  - Removed organizational context: `orgId`, `orgName`, `workspaceName`, `labels`
  - Removed execution metadata: `userName`, `commitId`, `scriptId`, `duration`, `exitStatus`, `success`, `manifest`, `nextflow`, `stats`, `errorMessage`, `errorReport`
  - Removed runtime paths: `projectDir`, `homeDir`, `launchDir`, `container`, `containerEngine`, `scriptFile`
  - Cleaned up `launch` block to remove internal fields: `sessionId`, `resumeDir`, `resumeCommitId`, `launchContainer`, `optimizationId`, `optimizationTargets`, `dateCreated`

- **Schema cleanup for `seqera_action`**: Removed 5 runtime and transient fields that should not be managed by Terraform:
  - Removed runtime event data: `event` (last event that triggered the action)
  - Removed timestamps: `lastSeen`, `dateCreated`, `lastUpdated`
  - Removed runtime label associations: `labels` (managed separately)

DEPRECATIONS:

The following items have been deprecated and will be getting replaced with suitable alternatives.

- **Deprecated Resource:** `seqera_compute_env` - Being replaced with compute environment specific resources (e.g., `seqera_aws_batch_ce`)
- **Deprecated Resource:** `seqera_credential` - Replaced with credential-specific resources (e.g., `seqera_aws_credential`, `seqera_github_credential`)
- **Deprecated Resource:** `seqera_aws_compute_env` - This has been renamed to `seqera_aws_batch_ce`
  - for users of `seqera_aws_compute_env` it is possible to use terraform state mv to `seqera_aws_batch_ce`

BUGFIXES:

- [85](https://github.com/seqeralabs/terraform-provider-seqera/issues/85) - CE region marked as optional
- [77](https://github.com/seqeralabs/terraform-provider-seqera/issues/77) - Value Conversion Erro
- [68](https://github.com/seqeralabs/terraform-provider-seqera/issues/68) - Terraform does not wait for a new TowerForge Compute Environment to become available
- [#83](https://github.com/seqeralabs/terraform-provider-seqera/issues/83) - Fixed `seqera_pipeline` resource to make `compute_env_id` and `work_dir` optional in the `launch` block
- [#81](https://github.com/seqeralabs/terraform-provider-seqera/issues/81) - Fixed `seqera_studios` documentation to clarify that `memory` is measured in megabytes (MB), not gigabytes
- [#67](https://github.com/seqeralabs/terraform-provider-seqera/issues/67)- Fixed field name typo: `nvnme_storage_enabled` renamed to `nvme_storage_enabled` in AWS Batch compute environments with automatic state migration
