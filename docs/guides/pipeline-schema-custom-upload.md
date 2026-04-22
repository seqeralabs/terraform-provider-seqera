---
page_title: "Upload a Custom Pipeline Schema"
subcategory: "Examples"
description: |-
  Create a Seqera pipeline schema from a nextflow_schema.json file and bind it to a pipeline so the Launchpad renders a curated parameter form.
---

# Upload a Custom Pipeline Schema

Seqera Platform uses a [pipeline schema](https://docs.seqera.io/platform-cloud/pipeline-schema/overview) — a JSON Schema document describing a pipeline's parameters — to validate inputs before launch and to render the Launchpad parameter form. By default, Seqera reads the `nextflow_schema.json` shipped with the pipeline's git repository. This guide walks through uploading a **custom** schema via Terraform so you can control which parameters the form exposes, pin validation rules that differ from the repo default, or surface a curated subset of fields for launch users.

~> **Note:** The `/pipeline-schemas` API is write-only. There is no endpoint to read, update, or delete a schema row by id. As a result, changes to `schema_content` force resource replacement (each apply POSTs a new row and rewires the pipeline), and `terraform destroy` leaves the previous schema row orphaned server-side. This is a backend constraint, not a provider limitation — see [Backend constraints and what they mean](#backend-constraints-and-what-they-mean) below.

## Prerequisites

- A Seqera Platform workspace and a compute environment in it.
- A pipeline schema file on disk. The usual source is a `nextflow_schema.json` from an nf-core pipeline or one generated with [`nf-core schema build`](https://nf-co.re/tools#build-a-pipeline-schema).

## Step 1: Create the schema resource

```terraform
terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
  }
}

variable "workspace_id" {
  type        = number
  description = "Seqera workspace ID that owns the pipeline."
}

resource "seqera_pipeline_schema" "rnaseq" {
  workspace_id   = var.workspace_id
  schema_content = file("${path.module}/nextflow_schema.json")
}
```

`schema_content` accepts the raw JSON Schema document as a string. `file()` is the typical way to load it from disk, but any expression that produces a string works — for example, `data.http.schema.response_body` to fetch from a URL at plan time, or `jsonencode()` of an HCL-native object.

## Step 2: Bind the schema to a pipeline

Reference the schema id from the pipeline's `launch.pipeline_schema_id`:

```terraform
resource "seqera_pipeline" "rnaseq" {
  workspace_id = var.workspace_id
  name         = "rnaseq-custom-schema"
  description  = "RNA-seq analysis with a curated parameter form"

  launch = {
    pipeline           = "https://github.com/nf-core/rnaseq"
    revision           = "3.14.0"
    compute_env_id     = seqera_compute_env.main.compute_env.id
    work_dir           = "s3://my-bucket/work"
    pipeline_schema_id = seqera_pipeline_schema.rnaseq.id
  }

  depends_on = [seqera_pipeline_schema.rnaseq]
}
```

The `pipeline_schema_id` reference already implies creation order, so the explicit `depends_on` is belt-and-suspenders: it keeps the relationship obvious and survives refactors that might pass the id through a `local` or variable and accidentally drop the implicit dependency. Terraform will create the schema first, then create the pipeline with `pipeline_schema_id` wired to the returned id. On subsequent launches, Seqera renders the Launchpad parameter form from the uploaded schema instead of the repository default.

## Step 3: Apply and verify

```shell
terraform apply
```

In the Seqera UI, open the pipeline and start a launch. The parameters form should reflect the fields defined in your uploaded schema — any parameters hidden by the schema will not appear in the launch form, and any validation rules (required, enums, regex patterns) are enforced before submission.

## Updating the schema

Because the API has no update endpoint, changing `schema_content` is a destroy + create:

```terraform
resource "seqera_pipeline_schema" "rnaseq" {
  workspace_id   = var.workspace_id
  schema_content = file("${path.module}/nextflow_schema.v2.json") # changed
}
```

`terraform plan` will show the resource being replaced. On apply, Terraform:

1. POSTs the new schema content to `/pipeline-schemas` — a new schema id.
2. Updates the `seqera_pipeline` resource to point `pipeline_schema_id` at the new id (because it depends on the schema's `id`).
3. Destroys the old `seqera_pipeline_schema` state entry. The previous schema row remains in the Seqera database, unreferenced.

The pipeline's configuration is coherent throughout — there is no window where the pipeline references a non-existent schema.

## Loading the schema from an nf-core pipeline checkout

If you track nf-core pipelines as git submodules or vendored directories, you can point `file()` at the checked-in `nextflow_schema.json`:

```terraform
resource "seqera_pipeline_schema" "rnaseq" {
  workspace_id   = var.workspace_id
  schema_content = file("${path.module}/../nf-core-pipelines/rnaseq/nextflow_schema.json")
}
```

This pairs well with `nf-core schema build` in CI: regenerate the schema, commit, `terraform apply` picks it up via the standard replace flow.

## Backend constraints and what they mean

The Seqera `/pipeline-schemas` API exposes only `POST`. There is no `GET /pipeline-schemas/{id}`, no `PUT`, and no `DELETE`. Consequences for Terraform state management:

- **Read is trusted from state.** Without a GET-by-id, the provider cannot reconcile the remote schema content against state. Schema rows are immutable server-side once created, so drift is not expected — but if it ever occurred the provider would silently trust state.
- **Update is replace.** Every `schema_content` change creates a new row. This is by design at the API layer, not a provider choice.
- **Delete is a no-op.** `terraform destroy` removes the resource from state but cannot delete the schema row in Seqera. Orphan rows accumulate in proportion to how often `schema_content` changes. This is accepted by the platform.

If the repo default schema is good enough for your use case, prefer leaving `launch.pipeline_schema_id` unset on `seqera_pipeline` — Seqera falls back to the repository's `nextflow_schema.json` automatically, and there is no orphan-row story to manage.

## Using a fetched schema

To download a schema from a URL at plan time — for example, to always take the latest `nextflow_schema.json` from a release tag — combine the `hashicorp/http` data source with the `seqera` provider:

```terraform
data "http" "rnaseq_schema" {
  url = "https://raw.githubusercontent.com/nf-core/rnaseq/3.14.0/nextflow_schema.json"
}

resource "seqera_pipeline_schema" "rnaseq" {
  workspace_id   = var.workspace_id
  schema_content = data.http.rnaseq_schema.response_body
}
```

Be aware that the fetched content is evaluated at every `terraform plan`, so if the upstream file changes the resource will be marked for replacement on the next apply. Pin the URL to a release tag or commit SHA to avoid surprise replacements.
