# `work_dir` Constraints for the Seqera Platform Terraform Provider

This document describes the API behavior of `work_dir` across Seqera Platform resources and the validation constraints the Terraform provider must enforce.

---

## Table of Contents

- [1. Overview](#1-overview)
- [2. Compute Environment Resource](#2-compute-environment-resource)
- [3. Pipeline Resource](#3-pipeline-resource)
- [4. Action Resource](#4-action-resource)
- [5. Workflow Launch (if applicable)](#5-workflow-launch-if-applicable)
- [6. Runtime Transformations](#6-runtime-transformations)
- [7. Source Code References](#7-source-code-references)

---

## 1. Overview

`work_dir` appears at three levels in the Seqera Platform API:

| Level                   | Description                                                                               |
| ----------------------- | ----------------------------------------------------------------------------------------- |
| **Compute Environment** | Default working directory stored in the CE config (e.g. `s3://my-bucket/work`)            |
| **Pipeline**            | Working directory on the pipeline's launch template — may override or omit the CE default |
| **Workflow Launch**     | The actual value used at execution time — must always resolve to non-null                 |

The critical behavioral difference is driven by **workspace visibility**: pipelines in **shared** workspaces can omit both the compute environment and `work_dir`, while pipelines in **private** workspaces (and the user/personal context) cannot.

---

## 2. Compute Environment Resource

### 2.1 On Creation

`work_dir` is a field inside the compute environment's `config` object. It is **not validated for format at creation time** by the API — format validation only occurs at workflow launch time. However, the Terraform provider should apply client-side validation per platform:

| Platform                                         | Valid `work_dir` Prefixes                                      |
| ------------------------------------------------ | -------------------------------------------------------------- |
| `aws-batch`                                      | `s3://` or `/`                                                 |
| `aws-cloud`                                      | `s3://` (Fusion required)                                      |
| `google-batch`                                   | `gs://` or `/`                                                 |
| `google-cloud` / `google-lifesciences`           | `gs://`                                                        |
| `azure-batch`                                    | `az://` or `/` (with Fusion: `az://` only)                     |
| `azure-cloud`                                    | `az://` (Fusion required)                                      |
| `seqeracompute-platform`                         | `s3://` (must be within the managed bucket)                    |
| `k8s-platform` / `eks-platform` / `gke-platform` | `/`, `s3://`, `gs://`, or `az://` (depends on storage backend) |
| `slurm-platform` / `lsf-platform` / other HPC    | `/` (local/NFS path)                                           |
| `local-platform`                                 | `/`                                                            |

### 2.2 On Update

**`work_dir` is immutable after creation.** The `PUT /compute-envs/{id}` endpoint only accepts `name` and `credentialsId`. To change `work_dir`, the CE must be destroyed and recreated.

**Terraform provider recommendation:** Mark `config.work_dir` as `ForceNew` so that a change triggers destroy + recreate.

### 2.3 Primary Compute Environment

Each workspace (or user, for personal context) can designate one CE as **primary** via `POST /compute-envs/{id}/primary`. The primary CE's `work_dir` serves as the fallback default when launching shared-workspace pipelines that have no CE bound. This is relevant context but not something the provider needs to enforce — the API handles the fallback.

---

## 3. Pipeline Resource

### 3.1 Core Rule: Workspace Visibility Determines Requirements

The API uses this logic when saving a pipeline's launch template:

```
isOptionalEnv = (workspace != null) AND (workspace.visibility != "PRIVATE")
```

Which produces these constraints:

| Context                                        | `compute_env_id`    | `work_dir`          |
| ---------------------------------------------- | ------------------- | ------------------- |
| **Personal workspace** (no `workspace_id`)     | **Required**        | **Required**        |
| **Private workspace** (`visibility = PRIVATE`) | **Required**        | **Required**        |
| **Shared workspace** (`visibility = SHARED`)   | Optional (nullable) | Optional (nullable) |

If `work_dir` is null in a private workspace or personal context, the API returns:

```
400 Bad Request: "Missing launching work directory"
```

If `compute_env_id` is null in the same contexts, the API returns:

```
400 Bad Request: "Missing launching environment profile"
```

### 3.2 No Auto-Defaulting from CE

When saving a pipeline template, the API does **not** auto-populate `work_dir` from the compute environment's config. The value from the request is stored as-is:

```groovy
launch.workDir = wfLaunchRequest.workDir  // direct assignment, no fallback
```

This means the Terraform provider **must** explicitly send `work_dir` when creating a pipeline in a private workspace — it will not be inferred from the compute environment.

> **Note:** Auto-defaulting from CE only happens in the quick-launch path (not used by Terraform) and at workflow launch time for shared workspaces.

### 3.3 Terraform Provider Validation

```
resource "seqera_pipeline" "example" {
  name         = "my-pipeline"
  workspace_id = seqera_workspace.ws.id

  launch {
    pipeline       = "https://github.com/org/repo"
    compute_env_id = seqera_compute_env.ce.id   # Required if workspace is PRIVATE or personal
    work_dir       = "s3://my-bucket/work"       # Required if workspace is PRIVATE or personal
  }
}
```

Recommended validation logic in the provider:

```
if workspace is null (personal context) OR workspace.visibility == "PRIVATE":
    compute_env_id → Required, error if empty
    work_dir       → Required, error if empty
elif workspace.visibility == "SHARED":
    compute_env_id → Optional
    work_dir       → Optional
```

### 3.4 Trailing Slash

The API strips trailing slashes from `work_dir` at launch time (`workDir.stripEnd('/')`). The provider should normalize this to avoid plan diffs — either strip on write or ignore trailing slashes in diff suppression.

---

## 4. Action Resource

Actions (automated pipeline triggers) **always require a compute environment**, regardless of workspace visibility. There is no shared/private distinction for actions.

When creating an action, `work_dir` follows the quick-launch defaulting path:

```groovy
result.workDir = launchOpts.workDir ?: env.config.workDir it

```

This means:

- If `work_dir` is provided, it is used as-is
- If `work_dir` is omitted, it defaults to the compute environment's `work_dir`
- The compute environment is always required

| Field            | Requirement                                       |
| ---------------- | ------------------------------------------------- |
| `compute_env_id` | **Always required**                               |
| `work_dir`       | Optional (defaults to CE's `work_dir` if omitted) |

When an action fires (via webhook or manual trigger), the stored `work_dir` from the action's launch template is used directly. The override config (webhook params) can only override `paramsText`, never `work_dir`.

---

## 5. Workflow Launch (if applicable)

If the Terraform provider ever supports triggering launches:

### 5.1 Quick Launch (no launch template ID)

`work_dir` is **always required** in the request body. No fallback logic.

### 5.2 Regular Launch from Private Workspace

`work_dir` is read-only for users with `launch` role (must match template value or be null). Users with `maintain`+ role can override it. Must resolve to non-null.

### 5.3 Regular Launch from Shared Workspace (with `sourceWorkspaceId`)

When the pipeline template has no bound CE:

- Both `compute_env_id` and `work_dir` become writable
- Fallback chain: request value -> template value -> primary CE's `work_dir` from the target workspace
- If user provides `compute_env_id`, it **must** be the target workspace's primary CE
- If user provides `work_dir`, it **must** match either the template's value or the primary CE's value

### 5.4 Resume Launch

`work_dir` is **immutable** — locked to the original workflow's resolved value. The only permitted CE change is one whose `work_dir` is a prefix of the existing workflow's `work_dir`.

---

## 6. Runtime Transformations

At execution time, platform providers may append a scratch sub-path to `work_dir` if only a bucket root is provided (no sub-path after the bucket name):

| Input                 | Transformed To                           |
| --------------------- | ---------------------------------------- |
| `s3://my-bucket`      | `s3://my-bucket/scratch/{workflowId}`    |
| `s3://my-bucket/work` | `s3://my-bucket/work` (unchanged)        |
| `gs://my-bucket`      | `gs://my-bucket/scratch/{workflowId}`    |
| `az://my-container`   | `az://my-container/scratch/{workflowId}` |

This transformation happens at launch time and is reflected in the workflow's `work_dir` response. The Terraform provider should **not** compare the stored pipeline `work_dir` against the runtime workflow `work_dir`.

---

## 7. Source Code References

| Behavior                                           | File                                                                          | Line(s)  |
| -------------------------------------------------- | ----------------------------------------------------------------------------- | -------- |
| `isOptionalEnv` for pipeline create                | `tower-enterprise/.../PipelineServiceImpl.groovy`                             | 663      |
| `isOptionalEnv` for pipeline update                | `tower-enterprise/.../PipelineServiceImpl.groovy`                             | 928      |
| Validation skip when optional                      | `tower-enterprise/.../LaunchServiceImpl.groovy`                               | 283-295  |
| Direct `workDir` assignment (no CE default)        | `tower-enterprise/.../LaunchServiceImpl.groovy`                               | 370      |
| Action `workDir` defaulting from CE                | `tower-enterprise/.../LaunchServiceImpl.groovy`                               | 108      |
| CE `workDir` immutable on update                   | `modules/platform-compute-env/api/.../UpdateComputeEnvRequest.groovy`         | 18-20    |
| Shared wsp launch fallback chain                   | `tower-enterprise/.../WorkflowLaunchServiceImpl.groovy`                       | 707-712  |
| Security checker (writable fields)                 | `modules/platform-launch/impl/.../LaunchDoesNotOverwriteFieldsChecker.groovy` | 91-99    |
| Primary CE constraint on shared launch             | `modules/platform-launch/impl/.../LaunchDoesNotOverwriteFieldsChecker.groovy` | 129-150  |
| Trailing slash strip                               | `tower-enterprise/.../WorkflowLaunchServiceImpl.groovy`                       | 350      |
| Scratch path append (AWS)                          | `tower-enterprise/.../AwsBatchPlatformProvider.groovy`                        | 709-710  |
| Scratch path append (GCP)                          | `tower-enterprise/.../GoogleCloudPlatformProvider.groovy`                     | 547-548  |
| Scratch path append (Azure)                        | `tower-enterprise/.../AzBatchPlatformProvider.groovy`                         | 180-181  |
| Resume `workDir` immutable                         | `tower-enterprise/.../WorkflowLaunchServiceImpl.groovy`                       | 738, 748 |
| Workspace visibility enum                          | `tower-core/.../Visibility.groovy`                                            | 17-34    |
| `getPrimary(workspace)`                            | `modules/platform-compute-env/impl/.../ComputeEnvServiceImpl.groovy`          | 116-118  |
| Per-platform `workDir` validation (AWS)            | `tower-enterprise/.../AwsBatchPlatformProvider.groovy`                        | 699-702  |
| Per-platform `workDir` validation (GCP)            | `tower-enterprise/.../GoogleCloudPlatformProvider.groovy`                     | 539-542  |
| Per-platform `workDir` validation (Azure)          | `tower-enterprise/.../AzBatchPlatformProvider.groovy`                         | 168-173  |
| Per-platform `workDir` validation (Seqera Compute) | `tower-enterprise/.../SeqeraComputePlatformProvider.groovy`                   | 286-292  |
