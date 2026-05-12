---
page_title: "Upgrading from v0.30.x to v0.40.0"
subcategory: "Upgrade Guides"
description: |-
  Step-by-step guide for upgrading the Seqera Terraform provider from v0.30.x to v0.40.0, covering automatic state upgraders, required HCL changes, deprecated fields, the new platform-specific compute environment resources, and Seqera Platform version compatibility.
---

# Upgrading from v0.30.x to v0.40.0

v0.40.0 introduces platform-specific compute environment resources, a deprecation of `seqera_aws_compute_env`, and one breaking schema change in Azure Batch (mitigated by an automatic state upgrader and a backwards-compat overlay). This guide walks through everything you need to know — what's automatic, what needs an HCL edit, and what's just nudging you toward better syntax.

## At a glance

| Change | Action required |
| --- | --- |
| Automatic state upgrades (typo fix, Azure Batch cleanup field) | None — runs on first `terraform plan` |
| `workspace_id` now required on all typed resources (drops user-context support) | Required HCL edit — add `workspace_id` to every typed resource, or migrate user-context resources to the generic `seqera_credential` |
| `seqera_aws_compute_env` → `seqera_aws_batch_ce` | Add a `moved {}` block (recommended); old resource still works with a deprecation warning |
| New first-class CE resources for AWS Batch / GCP Batch / GCP Cloud / Azure Batch / Azure Cloud | Optional — adopt via `moved {}` from `seqera_compute_env` if you want clearer typed schemas |
| `delete_jobs_on_completion` (string) → `delete_jobs_on_completion_enabled` (bool) on Azure Batch | Recommended HCL update; the string is still settable but deprecated |
| `ebs_auto_scale` / `ebs_block_size` deprecated (Fusion v2 incompatible) | Switch to `ebs_boot_size` if you want a larger root volume |
| Field renames `fusion2Enabled` → `enable_fusion`, `waveEnabled` → `enable_wave` | Use the new names if you're touching the resource |

## Before you start

1. **Pin and back up.** Pin your provider to v0.30.x in the lock file before you start, then commit and back up your `terraform.tfstate` (or remote state).
2. **Run a plan against v0.30.x first.** Make sure the existing config is clean — you don't want to debug an unrelated drift on top of an upgrade.
3. **Bump the provider.** Update the version constraint in `terraform { required_providers { ... } }` to `~> 0.40` (or pin to an exact 0.40.x once one is published) and run `terraform init -upgrade`.

## What runs automatically

### State upgraders

When you run `terraform plan` against v0.40.0 for the first time, the provider's `seqera_compute_env` state is upgraded from schema version 0 → 2 in two steps:

- **v0 → v1**: renames the typo'd `nvnme_storage_enabled` field to `nvme_storage_enabled` in the `aws_batch` config block.
- **v1 → v2**: in any `azure_batch` config block, if the legacy `delete_jobs_on_completion` string was set to a non-empty value (e.g. `"on_success"`), automatically sets `delete_jobs_on_completion_enabled = true`. This avoids a spurious `null → true` diff that would otherwise force the resource to be replaced.

Both steps only touch the platform sub-block they're meant for; other platforms in your state are untouched. You don't need to do anything to trigger this — it happens on the first plan/apply against v0.40.0.

### Server-managed forge audit fields removed from state

`forged_resources` and `deleted_resources` — server-recorded audit metadata on AWS Batch / AWS Cloud / Azure Cloud / GCP Cloud compute envs — are no longer in the provider schema. Existing state silently drops these attributes on first refresh; there's nothing to migrate.

### Field renames in generated code

`fusion2Enabled` and `waveEnabled` are renamed to `enable_fusion` and `enable_wave` respectively across all compute environment configs. State already uses the new names if your provider was at v0.30.x — the schema-side rename was already in place. Nothing for you to do here unless your HCL still spells the old names somewhere unusual; in that case Terraform will tell you.

## Required HCL edits

These are config edits Terraform won't do for you. Take them one at a time; each is independent.

### 1. `workspace_id` is now required on typed resources

All typed workspace-scoped resources now require `workspace_id`. This affects:

- All 12 typed credentials (`seqera_aws_credential`, `seqera_azure_credential`, `seqera_bitbucket_credential`, `seqera_codecommit_credential`, `seqera_container_registry_credential`, `seqera_gitea_credential`, `seqera_github_credential`, `seqera_gitlab_credential`, `seqera_google_credential`, `seqera_kubernetes_credential`, `seqera_ssh_credential`, `seqera_tower_agent_credential`)
- All 7 typed compute environments (`seqera_aws_batch_ce`, `seqera_aws_compute_env`, `seqera_azure_batch_ce`, `seqera_azure_cloud_ce`, `seqera_gcp_batch_ce`, `seqera_gcp_cloud_ce`, `seqera_seqera_compute_ce`)
- `seqera_pipeline`, `seqera_pipeline_secret`, `seqera_labels`, `seqera_data_studios`

User-context (personal workspace) resources are no longer supported on typed resources. If you were using a typed credential without `workspace_id`, you have two options:

- **Add `workspace_id`** if the resource belongs in an org workspace.
- **Move to the generic `seqera_credential`** — it remains the only resource that supports user-context credentials. See the [migrating from seqera_credential guide](migrating-from-seqera-credential.md) for the `import {}` / `removed {}` flow that preserves the Platform credential across the migration.

Typed credential resources also now support JSON-encoded import IDs that capture both `credentials_id` and `workspace_id`, so `terraform import` works correctly for workspace-scoped credentials (previously the import only set `credentials_id`, leaving `workspace_id` null and breaking the next refresh).

### 2. `seqera_aws_compute_env` → `seqera_aws_batch_ce` (recommended, optional)

`seqera_aws_compute_env` is now deprecated. It still works for v0.40.0 but emits a deprecation warning on every plan, and the registry doc page leads with a banner directing you here.

The new canonical resource is `seqera_aws_batch_ce`. The two resources share the same schema and API. Migrate without rebuilding by adding a `moved {}` block:

```terraform
moved {
  from = seqera_aws_compute_env.example
  to   = seqera_aws_batch_ce.example
}
```

Then rename the resource block itself:

```diff
- resource "seqera_aws_compute_env" "example" {
+ resource "seqera_aws_batch_ce" "example" {
    name           = "..."
    workspace_id   = ...
    platform       = "aws-batch"
    credentials_id = seqera_aws_credential.main.credentials_id
    config = {
      # unchanged
    }
  }
```

You can leave the `moved {}` block in place; once everyone on the team has updated, you can remove it.

### 3. Azure Batch `delete_jobs_on_completion` → `delete_jobs_on_completion_enabled`

In Seqera Platform v26.1, the legacy string field `delete_jobs_on_completion` (`"on_success"` / `"always"` / `"never"`) became read-only on the server and was replaced by three boolean fields:

- `delete_jobs_on_completion_enabled`
- `delete_pools_on_completion`
- `delete_tasks_on_completion`

The provider keeps the old string field settable so users on Platform v25.1 and earlier can still configure cleanup behaviour — but it's marked deprecated. When you're on v26.1+, switch your HCL:

```diff
  config = {
    azure_batch = {
-     delete_jobs_on_completion       = "on_success"
+     delete_jobs_on_completion_enabled = true
      # delete_pools_on_completion = true   # opt-in: also delete pools
      # delete_tasks_on_completion = true   # opt-in: also delete tasks
      ...
    }
  }
```

If your state was already populated under v0.30.x with a non-empty `delete_jobs_on_completion`, the automatic state upgrader (above) has already set `delete_jobs_on_completion_enabled = true` in state for you, so the diff stays clean once you make the HCL edit.

### 4. EBS auto-scale fields are deprecated

`ebs_auto_scale` and `ebs_block_size` in AWS Batch Forge configurations are now flagged deprecated in line with the upstream Seqera Platform — they aren't compatible with Fusion v2. If you set them, you'll see a deprecation warning. Migration is straightforward:

```diff
  forge = {
-   ebs_auto_scale = true
-   ebs_block_size = 100
+   ebs_boot_size  = 100      # size of the root volume in GB
    ...
  }
```

You can keep the deprecated fields temporarily; they continue to work.

## New resources you might want to adopt

v0.40.0 splits the per-cloud platforms out of the catch-all `seqera_compute_env` resource into first-class resources with typed schemas, dedicated registry doc pages, and platform-specific validators:

| New resource | Replaces using `seqera_compute_env` with… |
| --- | --- |
| `seqera_aws_batch_ce` | `platform = "aws-batch"` |
| `seqera_gcp_batch_ce` | `platform = "google-batch"` |
| `seqera_gcp_cloud_ce` | `platform = "google-cloud"` |
| `seqera_azure_batch_ce` | `platform = "azure-batch"` |
| `seqera_azure_cloud_ce` | `platform = "azure-cloud"` |

The catch-all `seqera_compute_env` still works for every platform, including ones that don't yet have a first-class resource (Kubernetes, EKS, GKE, Slurm, LSF, and others). Adopt the new resources at your own pace.

To migrate without rebuilding, use `moved {}` from `seqera_compute_env`:

```terraform
moved {
  from = seqera_compute_env.gcp_batch_example
  to   = seqera_gcp_batch_ce.gcp_batch_example
}
```

Each new resource accepts `moved {}` only from the catch-all `seqera_compute_env`. There is no cross-cloud or cross-platform move support — you can't move from `seqera_aws_batch_ce` to `seqera_gcp_batch_ce`, for example.

## Seqera Platform version compatibility

This is the matrix that matters for the `delete_jobs_on_completion` decision:

| Platform version | Old `delete_jobs_on_completion` (string) | New `delete_jobs_on_completion_enabled` (bool) etc. |
| --- | --- | --- |
| v25.1 and earlier | Source of truth — settable, no warning needed | Unknown to platform — don't use |
| v26.1+ | Read-only on the server; provider keeps it settable but marks deprecated | Authoritative, no warning |

The provider doesn't try to enforce one or the other based on platform version — that decision is yours, since the provider has no reliable way to detect the platform version mid-config. Pick the field that matches your platform.

## Provider behaviour changes worth knowing

These are improvements that don't require any action but are worth being aware of:

- **List-of-string fields accept unknown values from data sources.** `forge.subnets`, `forge.security_groups`, `forge.allow_buckets`, `forge.instance_types`, plus equivalents on AWS Cloud, GCP Batch, and Azure Batch configs, can now be fed directly from a Terraform data source — for example `subnets = data.aws_subnets.public.ids`. Previously this failed at plan time with `Received unknown value, however the target type cannot handle unknown values`.
- **Plan-time validators on AWS Batch.** v0.40.0 added validators that mirror the Seqera Platform UI: Fusion v2 requires Wave, Fast Instance Storage and Fusion Snapshots require Fusion v2, Fargate-for-head requires Fusion v2 + Spot, Graviton requires Fargate + Wave + Fusion v2. You'll see clear errors at `terraform plan` instead of failures at apply.
- **Credential improvements.** AWS credential supports `mode` (`keys` or `role`) with `external_id` and `use_external_id` for cross-account IAM. Google credential supports Workload Identity Federation. Azure credential supports Microsoft Entra ID (service principal). See the credential resource docs and the credential-specific guides.
- **Bitbucket app-password credentials.** `seqera_bitbucket_credential` now exposes the `password` field for the app-password auth mode (previously hidden). Git credential resource docs were refined to clarify auth-mode requirements across Bitbucket, CodeCommit, Gitea, GitHub, and GitLab.

## Verification

After upgrading and editing, run:

```sh
terraform init -upgrade
terraform plan
```

Expect to see:

- A "Resource Deprecated" or "Attribute Deprecated" warning if you're still using `seqera_aws_compute_env` or the legacy `delete_jobs_on_completion` string. These are non-fatal — plan still succeeds.
- No diff on `delete_jobs_on_completion_enabled` for resources that were previously using the old string (the state upgrader populated it).
- For each `moved {}` block you added: a one-time "moved … to …" line in the plan output, with no field-level diff.

If `terraform plan` errors or shows unexpected diffs, the most common causes are:

- HCL still references a field that's now read-only (e.g. you skipped Step 2 above).
- HCL still uses the pre-rename name for `enable_fusion` / `enable_wave`.
- A `moved {}` block targets the wrong destination resource type (only `seqera_compute_env` can be the source for the platform-specific `_ce` resources).

## Related guides

- [Migrating from `seqera_credential` to typed credential resources](migrating-from-seqera-credential.md)
- [AWS Credentials with IAM Role and External ID](aws-iam-role-with-external-id.md)
- [GCP Credentials with Workload Identity Federation](gcp-workload-identity-federation.md)
