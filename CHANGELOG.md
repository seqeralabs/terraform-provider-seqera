# v0.40.3

FEATURES:

- **New typed `seqera_slurm_ce` compute environment resource** for managing Slurm HPC clusters, continuing the split of platform-specific compute environments out of the generic `seqera_compute_env`. Seqera connects to the cluster's login/head node over SSH (via a `seqera_ssh_credential`) and submits the Nextflow head job to Slurm. Config fields (`work_dir`, `host_name`, `user_name`, `launch_dir`, `head_job_options`, queues, `pre_run_script`, etc.) are set at the resource root — there is no nested `config` block.

NOTES:

- **Import existing Slurm compute environments with `seqera_slurm_ce`, not the generic `seqera_compute_env`.** Importing a compute environment into the generic resource panics, because its polymorphic `compute_env` block is null when Terraform first reads state ([#226](https://github.com/seqeralabs/terraform-provider-seqera/issues/226)). The typed resources have a flat schema that imports cleanly; the `seqera_compute_env` docs now direct users to them.

# v0.40.2

BREAKING CHANGES:

- **`resume` removed** from `seqera_workflows`, `seqera_pipeline.launch`, and `seqera_action.launch`. As it was non-functional.

- **Read-only `workflow` run-report block removed** from `seqera_workflows`.

- **Read-only `launch.compute_env` object removed** from `seqera_pipeline` and `seqera_action`. It duplicated CE config already managed by the compute-env resources. `launch.compute_env_id` is unchanged.

No state migration is required — Terraform drops the old attributes on the next write.

BUGFIXES:

- **Removing a launch field from config now applies the removal** instead of silently planning "no changes". Previously, deleting an attribute such as `params_text` from a `seqera_pipeline` or `seqera_action` launch block left the old value server-side. Now:

  - `params_text`, `revision`, `config_text`, `tower_config`, `pre_run_script`, `post_run_script`, `main_script`, `entry_name`, `schema_name`, `run_name`, `head_job_cpus`, `head_job_memory_mb`, `pipeline_schema_id`, and `work_dir` clear on removal.
  - `pull_latest` and `stub_run` default to `false`; `config_profiles`, `user_secrets`, and `workspace_secrets` default to `[]`.

  The same defaults apply to the corresponding top-level fields on `seqera_workflows`.

ENHANCEMENTS:

- **Quieter plans.** Server-managed echo fields (`pipeline_id`, creator attribution, `launch_container`, `optimization_*`, `session_id`, etc.) no longer flip to "(known after apply)" on every in-place update, and a `seqera_workflows` replacement diff now shows the triggering line plus `workflow_id` instead of ~120 lines of resolved config.

UPGRADE NOTES:

- The first plan after upgrading shows a one-time in-place update on existing `seqera_workflows` resources, materialising the new defaults (`user_secrets`/`workspace_secrets` `null -> []`, `pull_latest`/`stub_run` `null -> false`). Applying it is a state-only write — no API calls, no runs launched or modified.

# v0.40.1

FEATURES:

- **New `seqera_github_app_credential` resource** for authenticating against a pre-existing GitHub App, as an alternative to a personal access token (`seqera_github_credential`).

# v0.40.0

FEATURES:

- **New typed compute environment resources.** Splits the platform-specific compute environments out of the catch-all `seqera_compute_env` into first-class resources with mode-specific schemas and validators: `seqera_gcp_batch_ce`, `seqera_gcp_cloud_ce`, `seqera_azure_batch_ce`, `seqera_azure_cloud_ce`, `seqera_aws_cloud_ce`. Existing `seqera_compute_env` deployments can migrate without re-creating the resource via a `moved {}` block:

  ```terraform
  moved {
    from = seqera_compute_env.example
    to   = seqera_gcp_batch_ce.example
  }
  ```

  Migrations are only supported from the generic `seqera_compute_env` (no cross-cloud or cross-platform moves).

- **Fine-grained access control (Cloud Pro / Enterprise v25.3+).** New `seqera_custom_role` resource + data source for org-scoped custom roles, plus `seqera_team` and `seqera_permissions` data sources. `seqera_workspace_participant` drops its role enum so custom roles are accepted alongside the predefined ones. See [docs/guides/fine-grained-access-control.md](docs/guides/fine-grained-access-control.md).

- **Pipeline version promotion and pinning.** New `seqera_pipeline_version` resource (rename in place, pin the platform default against out-of-band UI changes) and `seqera_pipeline_versions` data source. The resource does not create or delete versions — drafts appear automatically when a versionable field on `seqera_pipeline` changes. See [docs/guides/pipeline-versioning.md](docs/guides/pipeline-versioning.md).

- **New `seqera_pipeline_schema` resource for custom pipeline parameter schemas** — populates the Launchpad's custom parameters form. Server-side rows are immutable: content updates force replace, and `terraform destroy` is a no-op (no DELETE endpoint). See [docs/guides/pipeline-schema-custom-upload.md](docs/guides/pipeline-schema-custom-upload.md).

- **New typed Azure credential resources.** Splits the three Azure authentication modes previously conflated under `seqera_credential` into first-class resources: `seqera_azure_credential` (Batch shared key), `seqera_azure_entra_credential` (Batch with Entra service principal), and `seqera_azure_cloud_credential` (Cloud SingleVM with Entra service principal). The generic `seqera_credential` also accepts `provider_type = "azure-cloud"` and no longer crashes with "unknown after apply" when `keys.azure_cloud` is used. See [docs/guides/migrating-from-seqera-credential.md](docs/guides/migrating-from-seqera-credential.md).

- **AWS IAM role-based authentication on `seqera_aws_credential`**, with cross-account external ID support, alongside the existing access-key flow.

- **Google Workload Identity Federation on `seqera_google_credential`** — authenticate to GCP without storing a long-lived service account key. Recommended path for new deployments. See [GCP Credentials with Workload Identity Federation](docs/guides/gcp-workload-identity-federation.md).

- **Microsoft Entra ID (service principal) authentication on `seqera_azure_credential`**, alongside the existing shared-key flow.

ENHANCEMENTS:

- **Automatic retry on transient API failures.** All provider API calls now retry on connection errors, timeouts, HTTP 429, and HTTP 502/503/504, with exponential backoff (500ms → 30s, 5-minute cap). 4xx (other than 429) and HTTP 500 are not retried. Addresses prior `failure to invoke API` / `read: connection timed out` failures on `seqera_data_link`, `seqera_pipeline`, and `seqera_credential`.

- **Plan-time warning when pre/post-run scripts exceed 1024 bytes.** Seqera Cloud rejects scripts above this limit; Enterprise installs can raise it via platform config, so a warning (not an error) now fires at plan time on `pre_run_script` / `post_run_script` for compute environments, `seqera_pipeline.launch`, and `seqera_workflows`. Apply still proceeds.

- **Compute environment fields documented and named consistently across clouds.** Field descriptions, "requires replacement" annotations, and the `enable_fusion` / `enable_wave` naming (formerly `fusion2_enabled` / `wave_enabled`) now match across Google Cloud, Azure Cloud, AWS Cloud, and Google Cloud Batch.

- **SSH access on Studios** is now exposed in state, and `mount_data` has been restructured for stronger typing (the old string-list form is deprecated).

- **Plan-time validation of compute environment feature dependencies**, matching the Seqera Platform UI:

  - Fusion v2 requires Wave containers
  - Fast instance storage and Fusion Snapshots require Fusion v2
  - Fargate for head jobs requires Fusion v2 and Spot provisioning, and is not compatible with EFS or FSx
  - Graviton (ARM64) requires Fargate, Wave, and Fusion v2
  - Additional field-level validations for EBS, EFS, and DRAGEN dependencies
  - On AWS compute environments, `work_dir` must be a `s3://` URI with no trailing slash, and `sched_config` must be paired correctly with `sched_enabled` — both surfaced at plan time rather than as 4xx errors during apply

- **`.id` alias added to 20+ resources** for consistency with the Terraform convention `seqera_<resource>.<name>.id`. Both the new `.id` and the existing `{entity}_id` attribute hold the same value; existing HCL is unaffected. `seqera_custom_role` and `seqera_primary_compute_env` are intentionally excluded — their identities are composite or action-like.

DEPRECATIONS:

- **Azure Batch `delete_jobs_on_completion` is deprecated.** Replaced by three boolean fields: `delete_jobs_on_completion_enabled`, `delete_pools_on_completion`, `delete_tasks_on_completion`. The old string field still works (with a plan-time warning) for compatibility with Platform v25.1 and earlier; v26.1+ users should migrate (e.g. `delete_jobs_on_completion = "on_success"` → `delete_jobs_on_completion_enabled = true`). State is upgraded automatically when upgrading `seqera_compute_env` from v0.30.x; updating your config will not force a resource replacement.

- **`seqera_aws_compute_env`** is deprecated in favour of `seqera_aws_batch_ce`. Same schema and API; migrate via a `moved {}` block (see the resource docs). `terraform plan` surfaces a deprecation warning.

- **`ebs_auto_scale` and `ebs_block_size`** in AWS Batch Forge are deprecated — not compatible with Fusion v2. Use `ebs_boot_size` for the root volume.

BUGFIXES:

- **`seqera_action` updates no longer fail with HTTP 400 "Launch ID value can't change".** Originally reported on 0.30.4 as "actions don't update when secrets are involved" — in reality every `launch` mutation (compute env, params, secrets, revision) failed because the SDK stopped echoing back `launch.id` on PUT. The fix restores the round-trip; previously the only workaround was `terraform taint`.

- **Compute environments deleted in the UI no longer break `terraform plan`.** Previously, refreshing a CE that had been deleted outside Terraform failed with an unmarshal error; now `terraform plan` cleanly proposes to recreate it, and `terraform destroy` on an already-removed CE is a silent no-op. Affects every typed CE resource.

- **List fields now accept values from data sources.** `terraform plan` no longer fails with `Received unknown value, however the target type cannot handle unknown values` when feeding list-of-string fields from a data source (e.g. `subnets = data.aws_subnets.public.ids`). Affects `forge.subnets`, `forge.security_groups`, `forge.allow_buckets`, `forge.instance_types`, `allow_buckets`, `security_groups`, `compute_jobs_machine_type`, `network_tags`, and `container_reg_ids` across AWS/Azure/GCP compute environments.

- **Compute environment in-place updates.** Updates to `name`, `credentials_id`, or `description` no longer trigger a replace.

- **AWS Batch plan validation no longer crashes on DRAGEN, EFS, or Fusion Snapshots.** Field-dependency errors on `forge.dragen_*`, `forge.efs_create`, `forge.ebs_block_size`, and `fusion_snapshots` now surface as readable plan-time messages instead of path-expression crashes.

- **`pipeline` is now required on launch requests.** Previously optional on `seqera_pipeline`, `seqera_workflows`, and `seqera_action`; omitting it produced inconsistent 400 errors. Now enforced at plan time.

- **`workspace_id` is now read-only.** Assigned by the backend; marking it computed prevents accidental overrides.

- **No false drift on `enable_fusion` / `enable_wave` for Cloud CEs.** Fusion and Wave are hard requirements for Cloud compute environments and not user-configurable; the provider now omits them from the request and relies on the backend default, so they no longer appear as configurable fields.

- **Corrected EBS field descriptions on AWS compute environments.** `ebs_block_size` previously claimed to be the root volume — it's actually the auto-expandable scratch volume. `ebs_boot_size` (the real root volume) now has a description.

# v0.30.5

FIX:

- **Credentials** Fixed credential ID field mapping to correctly deserialize the `id` field from API responses across all credential resources.

# v0.30.4

ENHANCEMENTS:

- **Credentials** Updated provider to support `TOWER_ACCESS_TOKEN` Environment variable.

# v0.30.3

ENHANCEMENTS:

- **Credentials** Updated credentials to support [Write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments) please note these are only supported in Terraform 1.11 and later.

FIX:

- **Credentials** Removed inconsistent value of ID from credential resources resulting in an invalid result object after apply.

# v0.30.2

ENHANCEMENTS:

- **Refactored Member and Participant Resources** - Updated `seqera_organization_member`, `seqera_team_member`, and `seqera_workspace_participant` resources to use the new `PaginatedSearch` helper. This refactoring ensures consistent pagination behavior across all membership resources ensuring that in large organizations all resources work as intended.

# v0.30.1

FIX:

- **Compute Environment Lifecycle** - Improve the lifecycle management from Creating -> Created & Deleting -> Deleted for compatibility accross all Seqera Versions and handle both api.XXX & XXX/api endpoints.

- **Organization Member Role Management** - Fixed "Provider produced inconsistent result after apply" error when creating or updating members with non-default roles. The issue had two causes: (1) during creation, the desired role from the plan was being overwritten before the role update logic executed, and (2) during updates, eventual consistency in the API meant the list endpoint returned stale role data. The provider now saves the desired role before any API operations and preserves it after updates.

- **Workspace Participant Role Management** - Fixed "Provider produced inconsistent result after apply" error when creating or updating participants with non-default roles. Applied the same fixes as organization members: saving desired role before API operations and preserving it after updates to avoid eventual consistency issues.

- **Team Member Unnecessary Replacements** - Fixed issue where `seqera_team_member` resources were forcing unnecessary replacements when computed fields (role, avatar, name, etc.) changed externally. Added `UseStateForUnknown()` plan modifiers to all computed fields to prevent drift in read-only attributes from triggering resource recreation.

- **Computed Field Plan Modifiers** - Added `UseStateForUnknown()` plan modifiers to all computed fields in `seqera_organization_member`, `seqera_team_member`, and `seqera_workspace_participant` resources to prevent Terraform from forcing replacements when only read-only fields change.

- **Member Lookup Optimization** - Optimized all operations (Create, Read, Update) for `seqera_organization_member`, `seqera_team_member`, and `seqera_workspace_participant` resources to use ID-based filtering instead of email search when the ID is available. This eliminates unnecessary email lookup latency on every API call after initial creation, significantly improving performance and reducing API load. Email search is now only used during import operations when the ID is not yet known.

- **Organization Member Role Validation** - Fixed role validation for `seqera_organization_member` resource. The valid roles are now correctly set to: owner, member, view. Previously incorrectly allowed "collaborator" which is not a valid organization role.

# v0.30.0

FEATURES:

- **New Resource:** `seqera_workspace_participant` - Manage workspace participants with role assignment. Supports adding organization members to workspaces with roles: owner, admin, maintain, launch, or view.
- **New Resource:** `seqera_organization_member` - Manage organization members with role assignment. Supports adding users to organizations with roles: owner, member, or collaborator.
- **New Resource:** `seqera_team_member` - Manage team members. Supports adding organization members to teams for collective workspace access management.
- **New Resource:** `seqera_dataset_version` - Upload and manage dataset versions. Supports file uploads with header detection and SHA256 hash tracking for change detection.

- **New Data Source:** `seqera_organization_member` - Look up organization member by email. Returns member details including member_id, user_id, username, name, role, and avatar.
- **New Data Source:** `seqera_workspace` - Look up workspace by name. Returns workspace details including workspace_id, full_name, description, and visibility.
- **New Data Source:** `seqera_workspace_participant` - Look up workspace participant by email. Returns participant details including participant_id, member_id, username, name, and role.
- **New Data Source:** `seqera_pipeline` - Look up pipeline by name. Returns pipeline details including pipeline_id, description, repository, and creator information.
- **New Data Source:** `seqera_pipeline_secret` - Look up pipeline secret by name. Returns secret details including secret_id and timestamps.
- **New Data Source:** `seqera_organization` - Look up Organization by name. Returns Organization details including org_id, full_name, description

ENHANCEMENTS:

- **Resource Import Support**: All new resources support import via composite IDs:

  - `seqera_organization_member`: `org_id/email`
  - `seqera_workspace_participant`: `org_id/workspace_id/email`
  - `seqera_team_member`: `org_id/team_id/email`
  - `seqera_dataset_version`: `workspace_id/dataset_id/version`

- **Flexible User Identification**: `seqera_workspace_participant` and `seqera_team_member` resources accept either `member_id` or `email` for identifying users, with proper validation ensuring exactly one is specified.

- **File Change Detection**: `seqera_dataset_version` includes a computed `file_hash` attribute (SHA256) that triggers resource replacement when file content changes.

---

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
