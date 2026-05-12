---
page_title: "seqera_aws_cloud_ce Resource - terraform-provider-seqera"
subcategory: "Compute Environments"
description: |-
  Manage AWS Cloud compute environments in Seqera Platform.
  AWS Cloud compute environments execute Nextflow pipelines directly on
  EC2 instances managed by Seqera (rather than via AWS Batch). All
  configuration fields (region, work_dir, allow_buckets, instance_type,
  scripts, env, etc.) are identical between the two compute modes — the
  scheduler only adds the sched_config block; nothing else is
  hidden or unlocked.
  Two compute modes are supported, selected via sched_enabled.
  In Classic mode (sched_enabled = false, the default), worker
  fleet and spot-vs-on-demand strategy are managed by Tower Forge.
  Omit the sched_config block in this mode.
  In Seqera Intelligent Compute mode (Preview, sched_enabled = true),
  tasks are distributed across multiple EC2 instances with optimized
  scheduling and resource allocation. Set the sched_config block to
  choose the EC2 provisioning strategy and (optionally) restrict the
  instance-type catalog.
  Note: instance_type sets the head node EC2 type and applies in
  both modes (defaults: m5d.large, or m6gd.large when
  arm64_enabled = true).
  Backend feature flag. Enabling Seqera Intelligent Compute
  requires the SEQERA_SCHEDULER feature toggle on the target
  workspace/org. Without it, a create with sched_enabled = true
  returns HTTP 403. The toggle is controlled centrally — there is no
  API to flip it, so coordinate with the Platform team before applying.
---

# seqera_aws_cloud_ce (Resource)

Manage AWS Cloud compute environments in Seqera Platform.

AWS Cloud compute environments execute Nextflow pipelines directly on
EC2 instances managed by Seqera (rather than via AWS Batch). All
configuration fields (region, work_dir, allow_buckets, instance_type,
scripts, env, etc.) are identical between the two compute modes — the
scheduler only *adds* the `sched_config` block; nothing else is
hidden or unlocked.

Two compute modes are supported, selected via `sched_enabled`.

In **Classic** mode (`sched_enabled = false`, the default), worker
fleet and spot-vs-on-demand strategy are managed by Tower Forge.
Omit the `sched_config` block in this mode.

In **Seqera Intelligent Compute** mode (Preview, `sched_enabled = true`),
tasks are distributed across multiple EC2 instances with optimized
scheduling and resource allocation. Set the `sched_config` block to
choose the EC2 provisioning strategy and (optionally) restrict the
instance-type catalog.

Note: `instance_type` sets the **head node** EC2 type and applies in
both modes (defaults: `m5d.large`, or `m6gd.large` when
`arm64_enabled = true`).

**Backend feature flag.** Enabling Seqera Intelligent Compute
requires the `SEQERA_SCHEDULER` feature toggle on the target
workspace/org. Without it, a create with `sched_enabled = true`
returns HTTP 403. The toggle is controlled centrally — there is no
API to flip it, so coordinate with the Platform team before applying.

## Example Usage

```terraform
# Minimal AWS Cloud compute environment (Classic mode).
# Seqera picks the worker fleet automatically. Omit `sched_config` in this mode.
resource "seqera_aws_cloud_ce" "classic" {
  name           = "aws-cloud-classic"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-cloud"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region   = "us-west-1"
    work_dir = "s3://my-bucket/work"
  }
}
```

### Fusion Graviton

```terraform
# AWS Cloud (Classic mode) with Fusion v2, Wave, and Graviton (ARM64).
# Fusion v2 requires Wave; Graviton requires both.
resource "seqera_aws_cloud_ce" "fusion_graviton" {
  name           = "aws-cloud-fusion-graviton"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-cloud"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-west-1"
    work_dir      = "s3://my-bucket/work"
    enable_wave   = true
    enable_fusion = true
    arm64_enabled = true
    instance_type = "m7g.large" # Graviton head node
    ebs_boot_size = 100
  }
}
```

### Intelligent Compute

```terraform
# Seqera Intelligent Compute distributes tasks across multiple EC2 instances
# with optimised scheduling. See the resource docs for prerequisites and
# the SEQERA_SCHEDULER feature toggle requirement.
resource "seqera_aws_cloud_ce" "intelligent" {
  name           = "aws-cloud-intelligent"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "aws-cloud"
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-west-1"
    work_dir      = "s3://my-bucket/work"
    allow_buckets = ["s3://my-bucket-input", "s3://my-bucket-ref"]
    sched_enabled = true
    sched_config = {
      provisioning_model = "spotFirst" # spot | spotFirst | ondemand
      machine_types      = []          # empty = scheduler picks cost-optimal
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `config` (Attributes) Requires replacement if changed. (see [below for nested schema](#nestedatt--config))
- `credentials_id` (String) AWS credentials identifier. Requires replacement if changed.
- `name` (String) A unique name for this compute environment. Use only alphanumeric, dash, and underscore characters. Requires replacement if changed.
- `platform` (String) AWS platform type. must be "aws-cloud"; Requires replacement if changed.
- `workspace_id` (Number) Workspace numeric identifier. Requires replacement if changed.

### Optional

- `description` (String) Optional description of the compute environment. Requires replacement if changed.
- `label_ids` (List of Number) Requires replacement if changed.

### Read-Only

- `compute_env_id` (String) Compute environment string identifier
- `date_created` (String) Timestamp when the compute environment was created
- `deleted` (Boolean) Flag indicating if the compute environment has been deleted
- `id` (String) Unique identifier for the compute environment
- `last_updated` (String) Timestamp when the compute environment was last updated
- `last_used` (String) Timestamp when the compute environment was last used
- `org_id` (Number)
- `status` (String) Compute environment status

<a id="nestedatt--config"></a>
### Nested Schema for `config`

Required:

- `region` (String) AWS region where the compute environment will be created.
Examples: us-east-1, eu-west-1, ap-southeast-2
Requires replacement if changed.
- `work_dir` (String) S3 working directory for workflow execution. Must be a `s3://` URI in
the same region as the compute environment (e.g. `s3://my-bucket/work`).
Max 100 characters.
Requires replacement if changed.

Optional:

- `allow_buckets` (List of String) List of additional S3 bucket names that compute jobs are allowed to access.
The work directory bucket is automatically included.
Requires replacement if changed.
- `arm64_enabled` (Boolean) Enable ARM64 (Graviton) CPU architecture for compute instances.
When enabled, Graviton-based EC2 instances will be selected for cost savings.
Requires replacement if changed.
- `ebs_boot_size` (Number) Size of the boot disk (root volume) in GB for EC2 instances in this compute environment.
When using Fusion v2 without fast instance storage, this defaults to 100 GB with GP3 volume type.
Requires replacement if changed.
- `ec2_key_pair` (String) EC2 key pair name for SSH access to compute instances.
Key pair must exist in the specified region.
Requires replacement if changed.
- `enable_fusion` (Boolean) Allow access to your AWS S3-hosted data via the Fusion v2 virtual distributed file system,
speeding up most operations.

Requires `enable_wave = true`.
Requires replacement if changed.
- `enable_wave` (Boolean) Allow access to private container repositories and the provisioning of containers in your
Nextflow pipelines via the Wave containers service.

Required when `enable_fusion` is true.
Requires replacement if changed.
- `environment` (Attributes List) Requires replacement if changed. (see [below for nested schema](#nestedatt--config--environment))
- `gpu_enabled` (Boolean) Enable GPU support for compute instances.
When enabled, GPU-capable instance types will be selected.
Requires replacement if changed.
- `image_id` (String) Custom Amazon Machine Image (AMI) ID for compute instances.
If not specified, the default ECS-optimized AMI is used.
Requires replacement if changed.
- `instance_profile_arn` (String) IAM instance profile ARN for compute instances.
Format: arn:aws:iam::account-id:instance-profile/profile-name
Requires replacement if changed.
- `instance_type` (String) EC2 instance type for the compute environment (e.g., m5.xlarge, c5.2xlarge). Requires replacement if changed.
- `log_group` (String) CloudWatch Log group name for pipeline execution logs.
If specified, logs are sent to this existing log group instead of the default.
Requires replacement if changed.
- `nextflow_config` (String) Requires replacement if changed.
- `post_run_script` (String) Add a script that executes after all Nextflow processes have completed. See [Pre and post-run scripts](https://docs.seqera.io/platform-cloud/launch/advanced#pre-and-post-run-scripts). Requires replacement if changed.
- `pre_run_script` (String) Add a script that executes in the nf-launch script prior to invoking Nextflow processes. See [Pre and post-run scripts](https://docs.seqera.io/platform-cloud/launch/advanced#pre-and-post-run-scripts). Requires replacement if changed.
- `sched_config` (Attributes) Requires replacement if changed. (see [below for nested schema](#nestedatt--config--sched_config))
- `sched_enabled` (Boolean) Enable Seqera Intelligent Compute (Preview).
When `true`, tasks are distributed across multiple EC2 instances with
optimized scheduling and resource allocation, and `sched_config` is
required. When `false` (default), all tasks run on a single instance
(Basic mode) and `sched_config` must be omitted.

Setting this to `true` requires the `SEQERA_SCHEDULER` feature toggle
to be enabled on the target workspace/org; otherwise the API returns
HTTP 403.
Requires replacement if changed.
- `security_groups` (List of String) List of security group IDs to attach to compute instances.
Security groups must allow necessary network access.
Requires replacement if changed.
- `subnet_id` (String) Subnet ID where compute instances will be launched.
Must be in the same VPC and region as the compute environment.
Requires replacement if changed.

<a id="nestedatt--config--environment"></a>
### Nested Schema for `config.environment`

Optional:

- `compute` (Boolean) Whether this environment variable should be applied to compute/worker nodes.
At least one of 'head' or 'compute' must be set to true. Both can be true to target both environments.
Requires replacement if changed.
Default: false; Requires replacement if changed.
- `head` (Boolean) Whether this environment variable should be applied to the head/master node.
At least one of 'head' or 'compute' must be set to true. Both can be true to target both environments.
Requires replacement if changed.
Default: false; Requires replacement if changed.
- `name` (String) Requires replacement if changed.
- `value` (String) Requires replacement if changed.


<a id="nestedatt--config--sched_config"></a>
### Nested Schema for `config.sched_config`

Optional:

- `machine_types` (List of String) EC2 instance types eligible for Seqera Intelligent Compute nodes.
Leave empty (`[]`) to let the scheduler pick the most cost-optimal
types per task. When populated, the scheduler is restricted to this
whitelist; types outside the platform's filtered catalog for the
scheduler are accepted by the API but may produce warnings.
Requires replacement if changed.
- `provisioning_model` (String) EC2 provisioning strategy for Seqera Intelligent Compute nodes.
Case-sensitive — must be one of:
- `spotFirst` (default): try spot instances first, fall back to on-demand if capacity is unavailable. Recommended for cost.
- `spot`: spot instances only — lower cost, but jobs may be interrupted if capacity is reclaimed.
- `ondemand`: on-demand instances only — maximum reliability at a higher cost.

Note: `"onDemand"` / `"on-demand"` are rejected by the API.
Default: "spotFirst"; must be one of ["spot", "spotFirst", "ondemand"]; Requires replacement if changed.

## Import

Import is supported using the following syntax:

In Terraform v1.5.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `id` attribute, for example:

```terraform
import {
  to = seqera_aws_cloud_ce.my_seqera_aws_cloud_ce
  id = jsonencode({
    compute_env_id = "..."
    workspace_id   = 0
  })
}
```

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import seqera_aws_cloud_ce.my_seqera_aws_cloud_ce '{"compute_env_id": "...", "workspace_id": 0}'
```
