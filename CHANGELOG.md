# v0.25.1 (Unreleased)

FEATURES:

ENHANCEMENTS:

DEPRECATIONS:

BUGFIXES:

# v0.25.0

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
- Fixed field name typo: `nvnme_storage_enabled` renamed to `nvme_storage_enabled` in AWS Batch compute environments with automatic state migration

DEPRECATIONS:

The following items have been deprecated and will be getting replaced with suitable alternatives.

- **Deprecated Resource:** `seqera_compute_env` - Being replaced with compute environment specific resources (e.g., `seqera_aws_batch_ce`)
- **Deprecated Resource:** `seqera_credential` - Replaced with credential-specific resources (e.g., `seqera_aws_credential`, `seqera_github_credential`)
- **Deprecated Resource:** `seqera_aws_compute_env` - This has been renamed to `seqera_aws_batch_ce`
  - for users of `seqera_aws_compute_env` it is possible to use terraform state mv to `seqera_aws_batch_ce`

BUGFIXES:

https://github.com/seqeralabs/terraform-provider-seqera/issues/85 CE region marked as optional
https://github.com/seqeralabs/terraform-provider-seqera/issues/77 Value Conversion Erro
https://github.com/seqeralabs/terraform-provider-seqera/issues/68 Terraform does not wait for a new TowerForge Compute Environment to become available
