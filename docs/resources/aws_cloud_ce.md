---
page_title: "seqera_aws_cloud_ce Resource - terraform-provider-seqera"
subcategory: "Compute Environments"
description: |-
  Manage AWS Cloud compute environments in Seqera Platform.
  AWS Cloud compute environments execute Nextflow pipelines directly on
  EC2 instances managed by Seqera (rather than via AWS Batch). All
  configuration fields (region, work_dir, allow_buckets, instance_type,
  scripts, env, etc.) are identical between the two compute modes — the
  scheduler only adds the intelligent_compute_config block; nothing
  else is hidden or unlocked.
  Two compute modes are supported, selected via intelligent_compute_enabled.
  In Classic mode (intelligent_compute_enabled = false, the default),
  worker fleet and spot-vs-on-demand strategy are managed by Tower Forge.
  Omit the intelligent_compute_config block in this mode.
  In Seqera Intelligent Compute mode (Preview,
  intelligent_compute_enabled = true), tasks are distributed across
  multiple EC2 instances with optimized scheduling and resource
  allocation. The intelligent_compute_config block is optional —
  leave it null to accept the platform defaults, or set it to override
  the EC2 provisioning strategy or restrict the instance-type catalog.
  Note: instance_type sets the head node EC2 type and applies in
  both modes (defaults: m5d.large, or m6gd.large when
  arm64_enabled = true).
  Backend feature flag. Enabling Seqera Intelligent Compute
  requires the SEQERA_SCHEDULER feature toggle on the target
  workspace/org. Without it, a create with
  intelligent_compute_enabled = true returns HTTP 403. The toggle is
  controlled centrally — there is no API to flip it, so coordinate
  with the Platform team before applying.
---

# seqera_aws_cloud_ce (Resource)

Manage AWS Cloud compute environments in Seqera Platform.

AWS Cloud compute environments execute Nextflow pipelines directly on
EC2 instances managed by Seqera (rather than via AWS Batch). All
configuration fields (region, work_dir, allow_buckets, instance_type,
scripts, env, etc.) are identical between the two compute modes — the
scheduler only *adds* the `intelligent_compute_config` block; nothing
else is hidden or unlocked.

Two compute modes are supported, selected via `intelligent_compute_enabled`.

In **Classic** mode (`intelligent_compute_enabled = false`, the default),
worker fleet and spot-vs-on-demand strategy are managed by Tower Forge.
Omit the `intelligent_compute_config` block in this mode.

In **Seqera Intelligent Compute** mode (Preview,
`intelligent_compute_enabled = true`), tasks are distributed across
multiple EC2 instances with optimized scheduling and resource
allocation. The `intelligent_compute_config` block is optional —
leave it null to accept the platform defaults, or set it to override
the EC2 provisioning strategy or restrict the instance-type catalog.

Note: `instance_type` sets the **head node** EC2 type and applies in
both modes (defaults: `m5d.large`, or `m6gd.large` when
`arm64_enabled = true`).

**Backend feature flag.** Enabling Seqera Intelligent Compute
requires the `SEQERA_SCHEDULER` feature toggle on the target
workspace/org. Without it, a create with
`intelligent_compute_enabled = true` returns HTTP 403. The toggle is
controlled centrally — there is no API to flip it, so coordinate
with the Platform team before applying.

## Example Usage

```terraform
# Look up the target organization and workspace by name.
data "seqera_organization" "main" {
  name = "my-organization"
}

data "seqera_workspace" "main" {
  org_id = data.seqera_organization.main.org_id
  name   = "my-workspace"
}

# Minimal AWS Cloud compute environment (Classic mode).
# Seqera picks the worker fleet automatically.
#
# If you set `allow_buckets` explicitly, include the `work_dir` URI as the
# trailing entry — Seqera Forge implicitly appends it at CE-create time, and
# omitting it produces a forced-replacement diff on subsequent plans.
resource "seqera_aws_cloud_ce" "classic" {
  name           = "aws-cloud-classic"
  workspace_id   = data.seqera_workspace.main.id
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
  credentials_id = seqera_aws_credential.main.credentials_id

  config = {
    region        = "us-west-1"
    work_dir      = "s3://my-bucket/work"
    allow_buckets = ["s3://my-bucket-input", "s3://my-bucket-ref"]
    intelligent_compute_enabled = true
    intelligent_compute_config = {
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
- `credentials_id` (String) AWS credentials identifier
- `name` (String) A unique name for this compute environment. Use only alphanumeric, dash, and underscore characters.
- `workspace_id` (Number) Workspace numeric identifier. Requires replacement if changed.

### Optional

- `description` (String) Optional description of the compute environment
- `label_ids` (List of Number) Requires replacement if changed.

### Read-Only

- `compute_env_id` (String) Compute environment string identifier
- `date_created` (String) Timestamp when the compute environment was created
- `id` (String) Unique identifier for the compute environment
- `last_updated` (String) Timestamp when the compute environment was last updated
- `last_used` (String) Timestamp when the compute environment was last used
- `org_id` (Number)
- `platform` (String) AWS platform type. Always "aws-cloud" for this resource — set by the provider, not user-configurable. Default: "aws-cloud"
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
- `ebs_encrypted` (Boolean) When true, the boot EBS volume of provisioned instances is encrypted. Null/absent (the default) is treated as false — no encryption. Requires replacement if changed.
- `ebs_kms_key_id` (String) Optional KMS key ARN used to encrypt the boot EBS volume. Only applied when ebsEncrypted is true. When omitted, the account/region default EBS encryption key is used. Requires replacement if changed.
- `ec2_key_pair` (String) EC2 key pair name for SSH access to compute instances.
Key pair must exist in the specified region.
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
- `intelligent_compute_config` (Attributes) Requires replacement if changed. (see [below for nested schema](#nestedatt--config--intelligent_compute_config))
- `intelligent_compute_enabled` (Boolean) Enable Seqera Intelligent Compute (Preview).
When `true`, tasks are distributed across multiple EC2 instances with
optimized scheduling and resource allocation. When `false` (default),
all tasks run on a single instance (Classic mode).

`intelligent_compute_config` is optional in both modes: leave it null
to accept the platform defaults, or provide it (only when
`intelligent_compute_enabled = true`) to pin the provisioning strategy
or instance-type catalog.

Setting this to `true` requires the `SEQERA_SCHEDULER` feature toggle
to be enabled on the target workspace/org; otherwise the API returns
HTTP 403.
Requires replacement if changed.
- `log_group` (String) CloudWatch Log group name for pipeline execution logs.
If specified, logs are sent to this existing log group instead of the default.
Requires replacement if changed.
- `nextflow_config` (String) Requires replacement if changed.
- `post_run_script` (String) Add a script that executes after all Nextflow processes have completed. See [Pre and post-run scripts](https://docs.seqera.io/platform-cloud/launch/advanced#pre-and-post-run-scripts). Requires replacement if changed.
- `pre_run_script` (String) Add a script that executes in the nf-launch script prior to invoking Nextflow processes. See [Pre and post-run scripts](https://docs.seqera.io/platform-cloud/launch/advanced#pre-and-post-run-scripts). Requires replacement if changed.
- `security_groups` (List of String) List of security group IDs to attach to compute instances.
Security groups must allow necessary network access.
Requires replacement if changed.
- `subnet_id` (String, Deprecated) Subnet ID where compute instances will be launched.
Must be in the same VPC and region as the compute environment.
Requires replacement if changed.
- `subnet_ids` (List of String) Subnets to launch into. Basic uses the first; Intelligent Compute may use all. Requires replacement if changed.
- `vpc_id` (String) The VPC used to scope subnet and security-group selection. Requires replacement if changed.

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


<a id="nestedatt--config--intelligent_compute_config"></a>
### Nested Schema for `config.intelligent_compute_config`

Optional:

- `backend_strategy` (String) Backend used by Intelligent Compute to run tasks. 'ECS' (default) delegates task execution to AWS ECS; 'EC2' runs tasks directly on AWS EC2 instances. must be one of ["ECS", "EC2"]; Requires replacement if changed.
- `disk_allocation` (String) Requires replacement if changed.
- `fusion_snapshots` (Boolean) Enable Fusion snapshots so interrupted (e.g. spot-reclaimed) tasks can resume from a snapshot instead of restarting from scratch. Requires replacement if changed.
- `machine_types` (List of String) EC2 instance types eligible for Seqera Intelligent Compute nodes.
Leave empty (`[]`) to let the scheduler pick the most cost-optimal
types per task. When populated, the scheduler is restricted to this
whitelist; types outside the platform's filtered catalog for the
scheduler are accepted by the API but may produce warnings.
Requires replacement if changed.
- `nvme_enabled` (Boolean) When true, only use instance types providing local SSD (NVMe) storage. Maps to diskAllocation='nvme'. Requires replacement if changed.
- `pool` (Attributes) Warm-pool configuration. When present and enabled, the scheduler maintains a pool of idle VMs ready to absorb incoming tasks with sub-5s start latency. Requires replacement if changed. (see [below for nested schema](#nestedatt--config--intelligent_compute_config--pool))
- `prediction_model` (String) Resource-prediction model used by Intelligent Compute to size tasks. Suggested values: 'none' (default), 'qr/v1', 'qr/v2'. Any other string is accepted. Requires replacement if changed.
- `provisioning_model` (String) EC2 provisioning strategy for Seqera Intelligent Compute nodes.
Case-sensitive — must be one of:
- `spotFirst` (default): try spot instances first, fall back to on-demand if capacity is unavailable. Recommended for cost.
- `spot`: spot instances only — lower cost, but jobs may be interrupted if capacity is reclaimed.
- `ondemand`: on-demand instances only — maximum reliability at a higher cost.

Note: `"onDemand"` / `"on-demand"` are rejected by the API.
Default: "spotFirst"; must be one of ["spot", "spotFirst", "ondemand"]; Requires replacement if changed.

<a id="nestedatt--config--intelligent_compute_config--pool"></a>
### Nested Schema for `config.intelligent_compute_config.pool`

Optional:

- `desired_warm` (Number) Target number of idle VMs to keep warm. Bounds total warm-VM cost across all of this CE's pool clusters. Requires replacement if changed.
- `enabled` (Boolean) Whether the warm pool is active for this CE. When false, the scheduler will not maintain idle VMs. Requires replacement if changed.
- `scale_to_zero_secs` (Number) Seconds of inactivity after which the warm pool scales to zero. Set to 0 to never scale to zero. Requires replacement if changed.

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
