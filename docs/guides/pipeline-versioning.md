---
page_title: "Pin and Promote Pipeline Versions"
subcategory: "Examples"
description: |-
  Use Terraform to promote a known Seqera pipeline version to default and prevent it from drifting away — without trying to own the immutable audit trail of versions itself.
---

# Pin and Promote Pipeline Versions

Seqera Platform tracks every meaningful change to a pipeline as a **version** — an immutable, append-only audit trail. Day to day, users and maintainers create versions through the platform itself (publishing drafts, editing schemas, etc.). Where Terraform earns its keep is **promotion and pinning**: declaring "*this* named version is the one the Launchpad serves, and it stays that way until I change the config."

The two surfaces in this guide are deliberately narrow:

- **Promotion** — mark a chosen `version_id` as the pipeline default in one apply.
- **Drift correction** — if someone clicks "Make default" in the UI on a different version, the next plan flags it and apply puts the default back where the config says it should be.

Terraform is **not** the right tool for *creating* versions, *deleting* them, or owning every field on them — the platform's versioning model isn't shaped that way (see [Pipeline versioning docs](https://docs.seqera.io/platform-cloud/pipelines/versioning) for the full model). Day-to-day version creation happens through the platform UI or CLI. Terraform comes in afterwards to lock down which version is canonical.

~> **The provider does not create pipeline versions.** Versions are server-side artifacts: new **draft** versions are created automatically when a *versionable* field on the parent `seqera_pipeline` changes (schema parameters, schema selection, or any edit-form field except `name`, `image`, `description`, `labels`, and `resource_labels`). A maintainer **publishes** a draft by assigning it a name. The `seqera_pipeline_version` resource owns the `(name, is_default)` tuple of an *existing* version; it cannot conjure one. Use `seqera_pipeline` updates (or the UI/CLI) to spawn drafts, then use this resource to publish, rename, or promote them.

## Building blocks

| Surface | Wraps | Use it for |
|---|---|---|
| `data "seqera_pipeline_versions"` | `GET /pipelines/{pipelineId}/versions` | Discover `version_id`s, the current default, and version metadata (name, hash, timestamps, creator). |
| `resource "seqera_pipeline_version"` | `PUT /pipelines/{pipelineId}/versions/{versionId}/manage` | Rename a version, promote it to default, or publish a draft (assigning a name to a draft publishes it). |

## Constraints worth knowing upfront

- **Versions are immutable.** Neither drafts nor published versions can be deleted. `terraform destroy` on a `seqera_pipeline_version` is a no-op that releases Terraform's ownership; the version stays on the platform.
- **Exactly one default.** The platform refuses to leave a pipeline with zero defaults — the `/manage` endpoint returns `409 Cannot unset default flag on the current default version` if you try to demote the current default. To change defaults, *promote a different version* in the same plan rather than demoting the existing one.
- **Renames are in place.** Renaming a published version reuses the same `version_id` and `hash`. The docs describe this as "names reassigned to different draft versions" — the same primitive is reused for both rename and publish.

## The promotion + pin workflow

1. **Discover** the versions on a pipeline with the `seqera_pipeline_versions` data source.
2. **Choose** the one you want to be canonical — usually by name (`release-2024-Q4`, `v3.14.0`) so the config survives version IDs being regenerated in non-production workspaces.
3. **Pin** it with `seqera_pipeline_version { is_default = true }`.
4. Let Terraform's drift detection do the rest: anyone who promotes a different version out of band will see a non-empty `terraform plan` on the next run.

## Step 1: Discover the versions

```terraform
data "seqera_pipeline_versions" "rnaseq" {
  pipeline_id  = seqera_pipeline.rnaseq.pipeline_id
  workspace_id = var.workspace_id
}

output "rnaseq_versions" {
  value = [
    for v in data.seqera_pipeline_versions.rnaseq.versions : {
      id         = v.id
      name       = v.name
      is_default = v.is_default
      hash       = v.hash
    }
  ]
}
```

The data source returns one entry per version on the parent pipeline. Pass `is_published = true` (or `false`) to filter at the API boundary.

## Step 2: Promote a published version

The most common operation: pin a known version as default so it's the one launched from the Launchpad.

```terraform
locals {
  release_version_id = [
    for v in data.seqera_pipeline_versions.rnaseq.versions :
    v.id if v.name == "release-2024-Q4"
  ][0]
}

resource "seqera_pipeline_version" "release" {
  pipeline_id  = seqera_pipeline.rnaseq.pipeline_id
  workspace_id = var.workspace_id
  version_id   = local.release_version_id
  name         = "release-2024-Q4"
  is_default   = true
}
```

## Step 3: Publish a draft

When a versionable field on `seqera_pipeline` changes, the platform creates a draft automatically. Assign it a name through this resource to publish it:

```terraform
# A versionable change on the pipeline — main_script update — spawns a draft.
resource "seqera_pipeline" "rnaseq" {
  name         = "rnaseq"
  workspace_id = var.workspace_id

  launch = {
    pipeline       = "https://github.com/nf-core/rnaseq"
    compute_env_id = seqera_compute_env.aws.compute_env.id
    main_script    = "main.nf"
    revision       = "3.14.0" # change from 3.13.0 → spawns a draft
  }
}

data "seqera_pipeline_versions" "rnaseq_drafts" {
  pipeline_id  = seqera_pipeline.rnaseq.pipeline_id
  workspace_id = var.workspace_id
  is_published = false

  depends_on = [seqera_pipeline.rnaseq]
}

locals {
  draft_version_id = data.seqera_pipeline_versions.rnaseq_drafts.versions[0].id
}

resource "seqera_pipeline_version" "v3_14_0" {
  pipeline_id  = seqera_pipeline.rnaseq.pipeline_id
  workspace_id = var.workspace_id
  version_id   = local.draft_version_id
  name         = "v3.14.0"
  is_default   = true
}
```

Assigning `name = "v3.14.0"` to the draft publishes it; `is_default = true` promotes it in the same `/manage` call. Two-step in the UI; one apply here.

## Step 4: Rename a published version

Pure rename, no default change:

```terraform
resource "seqera_pipeline_version" "legacy" {
  pipeline_id  = seqera_pipeline.rnaseq.pipeline_id
  workspace_id = var.workspace_id
  version_id   = "5vh698AbE9wf8FET8GpEz"
  name         = "legacy-do-not-launch"
  is_default   = false
}
```

The `version_id`, `hash`, and creation timestamp are preserved. Only `name` and `last_updated` change.

## Common patterns

### Pin by name (variable-driven)

Reference the canonical version by name through a variable so promotion is a one-line config change reviewed in PR, not a UI click:

```terraform
variable "production_version_name" {
  description = "Name of the pipeline version blessed for production launches."
  type        = string
}

resource "seqera_pipeline_version" "prod_pinned" {
  pipeline_id  = seqera_pipeline.prod.pipeline_id
  workspace_id = var.workspace_id
  version_id = [
    for v in data.seqera_pipeline_versions.prod.versions :
    v.id if v.name == var.production_version_name
  ][0]
  name       = var.production_version_name
  is_default = true
}
```

### Audit all versions of a pipeline

```terraform
output "pipeline_audit" {
  value = {
    pipeline_id  = seqera_pipeline.rnaseq.pipeline_id
    default      = one([for v in data.seqera_pipeline_versions.rnaseq.versions : v.name if v.is_default])
    all_versions = [for v in data.seqera_pipeline_versions.rnaseq.versions : "${v.name} (${v.id})"]
  }
}
```

## Drift detection — the main reason to use Terraform here

`seqera_pipeline_version` lists versions on every Read and matches by `version_id`. The drift behaviour is intentional:

- **Someone clicks "Make default" in the UI on a different version.** Next `terraform plan` shows `is_default: false → true` on the configured version. Apply re-asserts via `/manage`. The "shadow" promotion is undone.
- **Someone renames the pinned version.** Same — `name: <new> → <pinned>` shows up on plan and is restored on apply.
- **The pinned version is removed from the platform.** The resource is removed from state and the next apply re-creates the assignment (which will fail loudly if the `version_id` no longer exists — surfacing the deletion).

This is what makes Terraform the right place to express "the production version of this pipeline is X": the platform allows any maintainer to flip the default at any time, and the config is the source of truth that catches and reverts those changes.

## Why not a `seqera_pipeline_draft_version` resource?

The platform does not expose a "create a draft version" API — drafts are emergent, appearing when a versionable field on the pipeline changes. Wrapping a non-existent endpoint would either be misleading (forcing pipeline re-create) or require this resource to mutate `seqera_pipeline` as a side effect, breaking resource isolation. The provider follows the API: `seqera_pipeline` owns pipeline-level fields; `seqera_pipeline_version` owns the `(name, is_default)` tuple on a discovered `version_id`.
