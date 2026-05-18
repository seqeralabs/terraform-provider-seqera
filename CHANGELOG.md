# v0.40.0

FEATURES:

- **New Resources:** Platform-specific Google Cloud, Azure, and AWS Cloud compute environment resources, mirroring `seqera_aws_batch_ce`. Splits the relevant platforms out of the catch-all `seqera_compute_env` resource into first-class resources with their own typed schemas, validators, and registry doc pages:

  - `seqera_gcp_batch_ce` — Google Cloud Batch (managed batch service) ([#112](https://github.com/seqeralabs/terraform-provider-seqera/issues/112))
  - `seqera_gcp_cloud_ce` — Google Cloud (Compute Engine VMs managed directly by Seqera) ([#113](https://github.com/seqeralabs/terraform-provider-seqera/issues/113))
  - `seqera_azure_batch_ce` — Azure Batch (managed pools) ([#116](https://github.com/seqeralabs/terraform-provider-seqera/issues/116))
  - `seqera_azure_cloud_ce` — Azure Cloud (VMs managed directly by Seqera) ([#115](https://github.com/seqeralabs/terraform-provider-seqera/issues/115))
  - `seqera_aws_cloud_ce` — AWS Cloud (EC2 instances managed directly by Seqera). Two compute modes via `sched_enabled`: Classic (default — Seqera picks the worker fleet) and Seqera Intelligent Compute (Preview — set `sched_config` with `provisioning_model` and an optional `machine_types` whitelist to control spot/on-demand strategy and the eligible instance types). Intelligent Compute requires the `SEQERA_SCHEDULER` feature toggle on the target workspace; without it, `terraform apply` will return 403. ([#114](https://github.com/seqeralabs/terraform-provider-seqera/issues/114))

  Each resource ships with a sidecar `MoveState` implementation, so existing `seqera_compute_env` deployments can migrate without re-creating the resource:

  ```terraform
  moved {
    from = seqera_compute_env.example
    to   = seqera_gcp_batch_ce.example
  }
  ```

  Migrations are only supported from the generic `seqera_compute_env` (no cross-cloud or cross-platform moves).

- **Fine-grained access control (Cloud Pro / Enterprise v25.3+):** New surfaces for managing custom roles and assigning them via existing team / workspace-participant resources ([#205](https://github.com/seqeralabs/terraform-provider-seqera/pull/205)):

  - `seqera_custom_role` (resource) — CRUD on org-scoped custom roles. `description` and `permissions` update in place; `name` change is force-new.
  - `seqera_custom_role` (data source) — resolves both custom and predefined roles; `is_predefined` distinguishes them.
  - `seqera_team` (data source) — by-name lookup, returns `team_id` and `members_count` for wiring teams into other resources without a hardcoded id.
  - `seqera_permissions` (data source) — live grant catalogue with optional category filter plus convenience `names` / `categories` lists; pairs with `lifecycle.precondition` to fail plan-time on invalid permission strings before any API call.
  - `seqera_workspace_participant` — drops the role enum so custom role names are accepted alongside predefined ones.

  See the new guide at [docs/guides/fine-grained-access-control.md](docs/guides/fine-grained-access-control.md) for end-to-end usage and three archetype role templates.

- **Pipeline version promotion and pinning.** New surfaces for declaring which version of a pipeline is the platform default and keeping it pinned against out-of-band UI changes:

  - `seqera_pipeline_versions` (data source) — lists versions on a pipeline; optional `is_published` filter forwards to the platform's `?isPublished` query parameter to scope to drafts or published versions.
  - `seqera_pipeline_version` (resource) — owns the `(name, is_default)` tuple of an existing version via `PUT /pipelines/{id}/versions/{versionId}/manage`. Renames are in place (same `version_id` and `hash`). `is_default = true` is always re-asserted on apply so out-of-band promotions in the UI are reverted at the next plan.

  The resource intentionally does **not** create or delete versions — the platform's audit trail is immutable and the API has no create-version or delete-version endpoint. Drafts appear automatically when a versionable field on `seqera_pipeline` changes; this resource publishes them (by assigning a name) and pins which one is default. See the new guide at [docs/guides/pipeline-versioning.md](docs/guides/pipeline-versioning.md).

- **New `seqera_pipeline_schema` resource for custom pipeline parameter schemas.** Wraps `POST /pipeline-schemas` and exposes the server-assigned `id` so it can be passed to `seqera_pipeline.launch.pipeline_schema_id`, populating the custom parameters form in the Launchpad. Read is trusted from state (rows are immutable server-side; the endpoint has no GET-by-id), updates to `schema_content` force replace, and `terraform destroy` is a no-op (no DELETE endpoint — the previous row is orphaned server-side). A new guide at [docs/guides/pipeline-schema-custom-upload.md](docs/guides/pipeline-schema-custom-upload.md) covers `file()` loading, URL fetch via `data.http`, and the nf-core checkout pattern. ([#184](https://github.com/seqeralabs/terraform-provider-seqera/pull/184))

- **New typed Azure credential resources** ([#204](https://github.com/seqeralabs/terraform-provider-seqera/pull/204)). Splits the three Azure authentication modes previously conflated under `seqera_credential` with `provider_type = "azure"` into first-class resources with mode-specific schemas:

  - `seqera_azure_credential` — Azure Batch shared key (`batch_name`, `storage_name`, `batch_key`, `storage_key`).
  - `seqera_azure_entra_credential` — Microsoft Entra service principal for Azure Batch (`batch_name`, `storage_name`, `tenant_id`, `client_id`, `client_secret`).
  - `seqera_azure_cloud_credential` — Microsoft Entra service principal for Azure Cloud SingleVM (`subscription_id`, `storage_name`, `tenant_id`, `client_id`, `client_secret`).

  The generic `seqera_credential` resource also accepts `provider_type = "azure-cloud"` and no longer crashes with "unknown after apply" when `keys.azure_cloud` is used (the `keys` block is no longer marked readonly). See [docs/guides/migrating-from-seqera-credential.md](docs/guides/migrating-from-seqera-credential.md) for migration from the generic resource.

ENHANCEMENTS:

- **SDK-level retry policy for transient API failures.** Every generated SDK call site now retries automatically on connection errors, timeouts, HTTP 429, and HTTP 502/503/504, with exponential backoff (500ms → 30s, 5-minute total cap). 4xx (other than 429) and HTTP 500 are *not* retried — they're either deterministic or could deepen partial-create orphans. Addresses prior reports of `failure to invoke API` / `read: connection timed out` failures on `seqera_data_link`, `seqera_pipeline`, and `seqera_credential` resources. ([#207](https://github.com/seqeralabs/terraform-provider-seqera/pull/207))

- **Plan-time warning when pre/post-run scripts exceed 1024 bytes.** Seqera Platform Cloud rejects pre/post-run scripts larger than 1024 bytes; Enterprise installs can raise the limit via platform configuration, so the provider can't know in advance which environment a config targets. A plan-time **Warning** (not an Error) now fires on every compute-environment `pre_run_script` / `post_run_script` above the limit, as well as on `seqera_pipeline.launch` and `seqera_workflows`. Apply still proceeds. ([#206](https://github.com/seqeralabs/terraform-provider-seqera/pull/206))

- **Compute environment fields are documented and behave consistently across clouds.** Field descriptions, "requires replacement" annotations, and the `enable_fusion` / `enable_wave` naming (formerly `fusion2_enabled` / `wave_enabled`) now match between Google Cloud, Azure Cloud, AWS Cloud, and Google Cloud Batch compute environments.

- **Compute Environments** - New configuration blocks added: `azure_cloud`, `google_cloud`.

- **Pipelines** - Added `version` block for pipeline versioning support and `pipeline_schema_id` field in the launch configuration.

- **Workflows** - Added `pipeline_schema_id` field for pipeline schema association.

- **Studios** - Added `ssh_details` read-only block with SSH connection information (`host`, `port`, `user`, `command`). Added `mount_data_v2` structured block (deprecating the `mount_data` string list) and `ssh_enabled` option in configuration.

- **Credentials** - AWS credential resource (`seqera_aws_credential`) now supports `mode` (`keys` or `role`), `external_id`, and `use_external_id` fields for IAM role-based authentication with cross-account external ID support.

- **Credentials** - Google credential resource (`seqera_google_credential`) now supports Workload Identity Federation via `workload_identity_provider`, `service_account_email`, and `token_audience` fields. WIF is the recommended path — no long-lived service account key is stored in the platform. See the new guide at [GCP Credentials with Workload Identity Federation](docs/guides/gcp-workload-identity-federation.md).

- **Credentials** - Azure credential resource (`seqera_azure_credential`) now supports Microsoft Entra ID (service principal) authentication via `tenant_id`, `client_id`, and `client_secret`, alongside the existing shared key flow. Removes the need for long-lived Azure access keys when using Batch Forge environments.

- **Validation** - Compute environment configuration now validates feature dependencies at plan time, matching the Seqera Platform UI. You'll get clear errors during `terraform plan` instead of unexpected failures at apply time.

  - Fusion v2 requires Wave containers
  - Fast instance storage and Fusion Snapshots require Fusion v2
  - Fargate for head jobs requires Fusion v2 and Spot provisioning, and is not compatible with EFS or FSx
  - Graviton (ARM64) requires Fargate, Wave, and Fusion v2
  - Additional field-level validations for EBS, EFS, and DRAGEN dependencies
  - `work_dir` on AWS compute environments must be a `s3://` URI with no trailing slash (catches typos like `"//s3"` at plan time, applies to `seqera_aws_batch_ce`, `seqera_aws_cloud_ce`, and the legacy `seqera_aws_compute_env`)
  - On `seqera_aws_cloud_ce`, `sched_config` must be set when `sched_enabled = true` and must be omitted when `sched_enabled = false` — surfaced at plan time rather than as a 4xx during apply

DEPRECATIONS:

- **Azure Batch `delete_jobs_on_completion` is deprecated.** Replaced by three boolean fields on Azure Batch compute configurations: `delete_jobs_on_completion_enabled`, `delete_pools_on_completion`, `delete_tasks_on_completion`. The old string field remains settable for backward compatibility with Seqera Platform v25.1 and earlier, but emits a deprecation warning at plan time. Users on Platform v26.1+ should migrate to the boolean fields.

  User migration:

  ```diff
  - delete_jobs_on_completion       = "on_success"
  + delete_jobs_on_completion_enabled = true
  ```

  State upgrade is handled automatically: when upgrading `seqera_compute_env` from v0.30.x, any non-empty `delete_jobs_on_completion` value is migrated to `delete_jobs_on_completion_enabled = true` in state, so updating your config doesn't force a resource replacement. Only the `azure_batch` config block is affected.

- **`seqera_aws_compute_env`** ([#187](https://github.com/seqeralabs/terraform-provider-seqera/issues/187)) - The `seqera_aws_compute_env` resource is now marked deprecated in favour of `seqera_aws_batch_ce`. The two resources share the same schema and API; `seqera_aws_batch_ce` is the canonical AWS Batch compute environment resource going forward. State can be migrated without re-creating the resource via a `moved {}` block — see the resource docs for an example. `terraform plan` will surface a deprecation warning, and the registry doc page now leads with a deprecation banner.

- **EBS Auto Scale** - `ebs_auto_scale` and `ebs_block_size` fields in AWS Batch Forge configuration are now marked as deprecated, matching the Seqera Platform documentation. These features are not compatible with Fusion v2. Use `ebs_boot_size` to configure a larger root volume instead.

BUGFIXES:

- **Compute environments deleted in the UI no longer break `terraform plan`.** Fixed an error when refreshing a compute environment that had been deleted outside of Terraform (e.g. through the Seqera Platform UI) — `terraform plan` would fail with an unmarshal error instead of cleanly proposing to recreate it. Affects every typed CE resource (`seqera_aws_batch_ce`, `seqera_aws_cloud_ce`, `seqera_azure_batch_ce`, `seqera_azure_cloud_ce`, `seqera_gcp_batch_ce`, `seqera_gcp_cloud_ce`, `seqera_managed_compute_ce`). `terraform destroy` on an already-removed CE is now also a silent no-op instead of an error.

- **AWS Batch plan validation no longer crashes on DRAGEN, EFS, or Fusion Snapshots.** Fixed `terraform validate` and `terraform plan` errors when using `forge.dragen_enabled`, `forge.dragen_ami_id`, `forge.dragen_instance_type`, `forge.efs_create`, `forge.ebs_block_size`, or `fusion_snapshots` on `seqera_aws_batch_ce`. The dependency rules between these fields are still enforced — they now run through plan-time validators that produce a readable error message instead of a path-expression crash.

- **List fields now accept values from data sources** ([#186](https://github.com/seqeralabs/terraform-provider-seqera/issues/186)). Fixed an error when driving list-of-string fields from a data source — for example, `subnets = data.aws_subnets.public.ids` or `network_tags = data.google_compute_network.x.tags`. Previously `terraform plan` failed with `Received unknown value, however the target type cannot handle unknown values`. Affects `forge.subnets`, `forge.security_groups`, `forge.allow_buckets`, `forge.instance_types` on AWS Batch resources, plus `allow_buckets` and `security_groups` on AWS Cloud, `compute_jobs_machine_type` and `network_tags` on GCP Batch, and `container_reg_ids` on Azure Batch.

- **Corrected EBS field descriptions on AWS compute environments** ([#159](https://github.com/seqeralabs/terraform-provider-seqera/issues/159)). `ebs_block_size` previously claimed to be the root volume size; it's actually the auto-expandable scratch volume. `ebs_boot_size` (the real root volume size) now has a description.

- **`pipeline` is now required on launch requests** ([#209](https://github.com/seqeralabs/terraform-provider-seqera/pull/209)). The nested `WorkflowLaunchRequest.pipeline` field was previously optional and shared across `seqera_pipeline`, `seqera_workflows`, and `seqera_action`. Omitting it produced inconsistent 400 errors from the backend (and an unhelpful message on actions); it's now enforced at plan time with a clear validation error.

- **`workspace_id` is now read-only** ([#210](https://github.com/seqeralabs/terraform-provider-seqera/pull/210)). The workspace identifier is assigned by the backend and should never be supplied by users — it's now marked computed to prevent accidental overrides.

- **Compute environment in-place updates** ([#211](https://github.com/seqeralabs/terraform-provider-seqera/pull/211)). Updates to `name`, `credentials_id`, or `description` on a compute environment no longer trigger a replace; the change is applied in place against the existing CE.

- **Cloud CEs ignore `enable_fusion` / `enable_wave`** ([#212](https://github.com/seqeralabs/terraform-provider-seqera/pull/212)). Fusion and Wave are hard requirements for Cloud compute environments and not user-configurable. The provider now omits these fields from the request rather than surfacing them as configurable, and relies on the backend to default them to enabled.

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
