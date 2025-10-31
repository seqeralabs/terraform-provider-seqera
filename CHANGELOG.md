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
