---
page_title: "seqera_pipeline_action Resource - terraform-provider-seqera"
subcategory: "Pipelines"
description: |-
  Manage an action that launches a pipeline run when it triggers.
  A pipeline action fires on a GitHub webhook (github), a call to
  its Seqera webhook (tower), a file arriving in a bucket
  (bucket), a schedule (cron) or the status of another pipeline's
  runs (pipeline_status), and launches a run with the settings in
  launch. GitHub webhooks require a GitHub account linked to the
  workspace. To run an agent instead, use seqera_agent_action.
  Migrating from seqera_action: an action that launches a pipeline
  can be moved here without recreating it, using a moved block
  (Terraform 1.8+).
---

# seqera_pipeline_action (Resource)

Manage an action that launches a pipeline run when it triggers.

A pipeline action fires on a GitHub webhook (`github`), a call to
its Seqera webhook (`tower`), a file arriving in a bucket
(`bucket`), a schedule (`cron`) or the status of another pipeline's
runs (`pipeline_status`), and launches a run with the settings in
`launch`. GitHub webhooks require a GitHub account linked to the
workspace. To run an agent instead, use `seqera_agent_action`.

Migrating from `seqera_action`: an action that launches a pipeline
can be moved here without recreating it, using a `moved` block
(Terraform 1.8+).

## Example Usage

```terraform
resource "seqera_pipeline_action" "tower_basic" {
  workspace_id = seqera_workspace.main.id
  name         = "api-triggered-pipeline"
  source       = "tower"

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
```

### Bucket Trigger

```terraform
# Launches a run when the marker file appears in the watched data link.
resource "seqera_pipeline_action" "bucket_trigger" {
  workspace_id = seqera_workspace.main.id
  name         = "run-on-upload"
  source       = "bucket"

  bucket = {
    data_link_id = seqera_data_link.inputs.data_link_id
    marker_file  = "ready.txt"
  }

  launch = {
    pipeline       = "https://github.com/nf-core/rnaseq"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"

    params_text = jsonencode({
      input  = "s3://my-bucket/inputs/samplesheet.csv"
      outdir = "s3://my-bucket/results"
    })
  }
}
```

### Cron Schedule

```terraform
# Launches a run every night at 02:00 UTC.
resource "seqera_pipeline_action" "nightly" {
  workspace_id = seqera_workspace.main.id
  name         = "nightly-run"
  source       = "cron"

  cron = {
    expression = "0 2 * * *"
    timezone   = "UTC"
  }

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
```

### Github Webhook

```terraform
# IMPORTANT: GitHub webhook actions require a logged in GitHub account associated
# with the user creating the action.
resource "seqera_pipeline_action" "github_webhook" {
  workspace_id = seqera_workspace.main.id
  name         = "github-push-trigger"
  source       = "github"

  launch = {
    pipeline       = "https://github.com/myorg/my-pipeline"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"

    config_profiles = ["docker", "aws"]

    params_text = jsonencode({
      input  = "s3://my-bucket/input.csv"
      output = "s3://my-bucket/results"
    })
  }
}
```

### Moved From Action

```terraform
# Moves a pipeline action managed by seqera_action to seqera_pipeline_action
# without recreating it. Requires Terraform 1.8 or later. Replace the
# seqera_action block with a seqera_pipeline_action one with the same
# arguments, dropping response_type if it was set.
moved {
  from = seqera_action.nightly
  to   = seqera_pipeline_action.nightly
}

resource "seqera_pipeline_action" "nightly" {
  workspace_id = seqera_workspace.main.id
  name         = "nightly-run"
  source       = "cron"

  cron = {
    expression = "0 2 * * *"
    timezone   = "UTC"
  }

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
```

### Pipeline Status Trigger

```terraform
# Launches a downstream run each time a run of the watched pipeline succeeds.
# The watched pipeline is set in pipeline_status; the run the action starts is
# set in launch.
resource "seqera_pipeline_action" "on_upstream_success" {
  workspace_id = seqera_workspace.main.id
  name         = "run-after-upstream"
  source       = "pipeline_status"

  pipeline_status = {
    pipeline_id = seqera_pipeline.upstream.pipeline_id
    run_status  = "SUCCEEDED"
  }

  launch = {
    pipeline       = "https://github.com/myorg/downstream-pipeline"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
```

### Target Pipeline

```terraform
# Links the action to a Launchpad pipeline so each run is recorded as a run of
# that pipeline. The run itself still uses the settings in launch.
resource "seqera_pipeline_action" "linked" {
  workspace_id = seqera_workspace.main.id
  name         = "linked-nightly-run"
  source       = "cron"

  cron = {
    expression = "0 2 * * *"
    timezone   = "UTC"
  }

  pipeline = {
    target_pipeline_id = seqera_pipeline.hello.pipeline_id
  }

  launch = {
    pipeline       = "https://github.com/nextflow-io/hello"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/work"
    revision       = "master"
  }
}
```

### Tower Advanced

```terraform
resource "seqera_pipeline_action" "tower_advanced" {
  workspace_id = seqera_workspace.main.id
  name         = "production-pipeline"
  source       = "tower"

  launch = {
    pipeline       = "https://github.com/myorg/production-pipeline"
    compute_env_id = seqera_compute_env.aws.id
    work_dir       = "s3://my-bucket/production/work"
    revision       = "master"
    syntax_parser  = "v2"

    params_text = jsonencode({
      input_data  = "s3://my-bucket/input/data.csv"
      output_dir  = "s3://my-bucket/results"
      sample_size = 1000
    })

    config_text = <<-EOT
      process {
        executor = 'awsbatch'
        queue    = 'my-production-queue'
        memory   = '8 GB'
        cpus     = 4
      }
    EOT

    pre_run_script = <<-EOT
      #!/bin/bash
      echo "Starting production pipeline run"
      aws s3 sync s3://my-bucket/reference ./reference
    EOT

    post_run_script = <<-EOT
      #!/bin/bash
      echo "Workflow completed"
      aws s3 sync ./results s3://my-bucket/results
    EOT

    pull_latest = true
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `launch` (Attributes) Launch settings for the run the action starts. The whole block is sent on every update. (see [below for nested schema](#nestedatt--launch))
- `name` (String) Name of the action.
- `workspace_id` (Number) Workspace numeric identifier. Requires replacement if changed.

### Optional

- `bucket` (Attributes) Bucket trigger. Fires when the marker file appears under the data link. Required when `source` is `bucket`; must not be set otherwise. (see [below for nested schema](#nestedatt--bucket))
- `cron` (Attributes) Cron trigger. Fires on a schedule. Required when `source` is `cron`; must not be set otherwise. Set either `expression` or `preset`. (see [below for nested schema](#nestedatt--cron))
- `pipeline` (Attributes) Links the action to an existing Launchpad pipeline and version, so each run is recorded as a run of that pipeline. The run itself uses the settings in `launch`. On a `pipeline_status` action, the pipeline to watch is set in `pipeline_status`, not here. (see [below for nested schema](#nestedatt--pipeline))
- `pipeline_status` (Attributes) Pipeline-status trigger. Fires when a run of the watched pipeline reaches `run_status`. Required when `source` is `pipeline_status`; must not be set otherwise. (see [below for nested schema](#nestedatt--pipeline_status))
- `source` (String) Event source that triggers the action. `tower` (the default) is a webhook called programmatically; `github` requires a GitHub account linked to the workspace. Default: "tower"; must be one of ["github", "tower", "bucket", "cron", "pipeline_status"]; Requires replacement if changed.

### Read-Only

- `error` (String) Error reported by the Platform for the action, if any.
- `hook_id` (String) Identifier of the webhook that triggers a `github` or `tower` action.
- `hook_url` (String) URL of the webhook that triggers a `github` or `tower` action.
- `id` (String) Action identifier.
- `next_execution` (String) Next scheduled run of a `cron` action.
- `status` (String) Current status of the action.

<a id="nestedatt--launch"></a>
### Nested Schema for `launch`

Required:

- `pipeline` (String)

Optional:

- `compute_env_id` (String)
- `config_profiles` (List of String) Default: []
- `config_text` (String) Nextflow configuration text
- `entry_name` (String) Entry workflow name
- `fusion_snapshots` (Boolean) Whether to enable Fusion snapshots for this workflow. Can only be enabled when the compute environment runs Wave, Fusion v2 and Spot instances. When null, inherits the source pipeline or run, then the compute environment setting.
- `fusion_version` (String) Fusion version to run this workflow with; must exist in the system catalog and be compatible with the deployed services. Applies only when the compute environment enables Wave and Fusion v2 and overrideNextflowFusionVersion is enabled. When null or empty, inherits the source pipeline or run, then the compute environment, then the catalog default.
- `head_job_cpus` (Number) Head job CPU allocation
- `head_job_memory_mb` (Number) Head job memory allocation in MB
- `label_ids` (List of Number)
- `main_script` (String) Main script path
- `nextflow_version` (String) Nextflow release version to run this workflow with; must exist in the system catalog and satisfy the minimum configured for the compute environment's type.
- `output_dir` (String) Per-run output directory, passed to Nextflow as `-output-dir`. Requires
Nextflow 24.10.0 or later and the workflow outputs syntax.
- `override_nextflow_fusion_version` (Boolean) Whether to run this workflow with fusionVersion instead of the Fusion version provided by the Nextflow launcher container. Applies only when the compute environment enables Wave and Fusion v2. When null, inherits the source pipeline or run, then the compute environment setting.
- `params_text` (String) Pipeline parameters text
- `pipeline_schema_id` (Number)
- `post_run_script` (String) Script to run after pipeline execution
- `pre_run_script` (String) Script to run before pipeline execution
- `pull_latest` (Boolean) Default: false
- `revision` (String) Pipeline revision
- `run_name` (String) Custom run name
- `schema_name` (String) Pipeline schema name
- `stub_run` (Boolean) Default: false
- `syntax_parser` (String) must be one of ["v1", "v2"]
- `tower_config` (String) Tower-specific configuration
- `user_secrets` (List of String) Default: []
- `work_dir` (String) Working directory for pipeline execution. Must start with a valid cloud storage prefix (s3://, gs://, az://) or be an absolute local path (/). Do not include a trailing slash — the API strips trailing slashes at launch time, which causes plan diffs. Required for pipelines in private workspaces and personal context; optional for shared workspaces. You can reference the work_dir from your compute environment instead of duplicating the value, e.g. seqera_compute_env.my_ce.compute_env.config.aws_batch.work_dir or seqera_aws_batch_compute_env.my_ce.config.work_dir.
- `workspace_secrets` (List of String) Default: []

Read-Only:

- `id` (String) Server-generated launch identifier; echoed back on Update.
- `launch_container` (String)
- `optimization_id` (String) Optimization profile ID
- `optimization_targets` (String) Optimization targets
- `resume_launch_id` (String) Launch ID to resume from
- `session_id` (String) Session ID for resuming
- `workspace_id` (Number)


<a id="nestedatt--bucket"></a>
### Nested Schema for `bucket`

Required:

- `data_link_id` (String) ID of the data link (bucket) to watch. Requires replacement if changed.

Optional:

- `events` (List of String) Bucket events that fire the action.
- `marker_file` (String) Name of the file whose arrival fires the action.


<a id="nestedatt--cron"></a>
### Nested Schema for `cron`

Optional:

- `expression` (String) Cron expression for the schedule, for example `0 2 * * *`.
- `preset` (String) Named schedule preset, used instead of `expression`.
- `timezone` (String) IANA timezone the schedule runs in, for example `Europe/Madrid`.


<a id="nestedatt--pipeline"></a>
### Nested Schema for `pipeline`

Optional:

- `target_pipeline_id` (Number) ID of the Launchpad pipeline to link.
- `target_pipeline_version_id` (String) ID of the pipeline version to link.


<a id="nestedatt--pipeline_status"></a>
### Nested Schema for `pipeline_status`

Required:

- `pipeline_id` (Number) ID of the Launchpad pipeline to watch.
- `run_status` (String) Run status that fires the action. must be one of ["SUBMITTED", "RUNNING", "SUCCEEDED", "FAILED", "CANCELLED", "UNKNOWN"]

## Import

Import is supported using the following syntax:

In Terraform v1.5.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `id` attribute, for example:

```terraform
import {
  to = seqera_pipeline_action.my_seqera_pipeline_action
  id = jsonencode({
    id           = "..."
    workspace_id = 0
  })
}
```

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import seqera_pipeline_action.my_seqera_pipeline_action '{"id": "...", "workspace_id": 0}'
```
