---
page_title: "Waiting for Compute Environment Availability"
subcategory: "Guides"
description: |-
  Guide on handling compute environment provisioning delays using time_sleep resource
---

# Waiting for Compute Environment Availability

## Overview

When creating compute environments (especially with TowerForge or auto-provisioning), the Seqera API returns success immediately, but the actual infrastructure provisioning can take several minutes to complete. If you try to use the compute environment in dependent resources (like pipeline launches) before it reaches the `AVAILABLE` state, you may encounter errors.

## The Problem

**Symptom**: You may see errors like:
```
ERROR: Compute environment 'ABCDEF1234556' not found at 0987654321 workspace
```

This happens because:
1. Terraform creates the compute environment resource successfully
2. The provider returns immediately (after ~1 second)
3. The actual cloud infrastructure is still being provisioned (takes 2-5 minutes)
4. Dependent resources try to use the compute environment before it's ready

## Solution: Using time_sleep Resource

Use the `time_sleep` resource from the `hashicorp/time` provider to add a delay between compute environment creation and usage.

### Basic Example

```terraform
terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

# Create compute environment
resource "seqera_compute_env" "my_env" {
  compute_env = {
    name           = "my-forge-environment"
    platform       = "aws-batch"
    credentials_id = var.aws_credentials_id
    config = {
      aws_batch = {
        region = "us-east-1"
        work_dir = "s3://my-bucket/work"
        forge = {
          type     = "SPOT"
          min_cpus = 0
          max_cpus = 256
        }
      }
    }
  }
  workspace_id = var.workspace_id
}

# Wait for provisioning to complete
resource "time_sleep" "wait_for_compute_env" {
  depends_on      = [seqera_compute_env.my_env]
  create_duration = "4m"
}

# Use compute environment in dependent resources
resource "seqera_pipeline" "my_pipeline" {
  depends_on = [time_sleep.wait_for_compute_env]

  pipeline = {
    name = "my-pipeline"
    launch = {
      pipeline       = "nf-core/rnaseq"
      compute_env_id = seqera_compute_env.my_env.compute_env_id
      work_dir       = "s3://my-bucket/work"
    }
  }
  workspace_id = var.workspace_id
}
```

## Recommended Wait Times

The appropriate wait time depends on your compute environment type and configuration:

| Platform | Configuration | Recommended Wait Time |
|----------|--------------|----------------------|
| AWS Batch | With TowerForge | 3-5 minutes |
| AWS Batch | Existing queue | 30-60 seconds |
| AWS Cloud | - | 2-3 minutes |
| Google Life Sciences | - | 2-3 minutes |
| Google Batch | - | 2-3 minutes |
| Azure Batch | With Forge | 3-5 minutes |
| Azure Batch | Existing pool | 1-2 minutes |
| EKS | - | 1-2 minutes |
| GKE | - | 1-2 minutes |
| Kubernetes (existing) | - | 30-60 seconds |
| Grid schedulers (Slurm/LSF/UGE) | - | 30 seconds |
| Local | - | 10 seconds |

**Note**: These are conservative estimates. You may need to adjust based on your specific environment and cloud provider response times.

## Best Practices

### 1. Add Triggers for Recreating Wait

If the compute environment is recreated, ensure the wait is also recreated:

```terraform
resource "time_sleep" "wait_for_compute_env" {
  depends_on      = [seqera_compute_env.my_env]
  create_duration = "4m"

  triggers = {
    compute_env_id = seqera_compute_env.my_env.compute_env_id
  }
}
```

### 2. Document the Wait Reason

Add comments to explain why the wait is necessary:

```terraform
# Wait 4 minutes for AWS Batch Forge provisioning to complete.
# TowerForge needs to create VPC resources, compute environments,
# and job queues before the environment is usable.
resource "time_sleep" "wait_for_compute_env" {
  depends_on      = [seqera_compute_env.my_env]
  create_duration = "4m"
}
```

### 3. Use Variables for Wait Duration

Make the wait time configurable:

```terraform
variable "compute_env_wait_duration" {
  description = "Time to wait for compute environment to become available"
  type        = string
  default     = "4m"
}

resource "time_sleep" "wait_for_compute_env" {
  depends_on      = [seqera_compute_env.my_env]
  create_duration = var.compute_env_wait_duration
}
```

### 4. Apply to All Dependent Resources

Ensure ALL resources that depend on the compute environment also depend on the time_sleep:

```terraform
# Correct: Both pipeline and workflow depend on time_sleep
resource "seqera_pipeline" "pipeline" {
  depends_on = [time_sleep.wait_for_compute_env]
  # ...
}

resource "seqera_workflows" "workflow" {
  depends_on = [time_sleep.wait_for_compute_env]
  # ...
}

# Incorrect: Direct dependency on compute environment
resource "seqera_pipeline" "pipeline" {
  depends_on = [seqera_compute_env.my_env]  # Don't do this!
  # ...
}
```

## Module Example

When using modules, pass the time_sleep resource as an output:

```terraform
# modules/compute-env/main.tf
resource "seqera_compute_env" "this" {
  # ... configuration
}

resource "time_sleep" "wait" {
  depends_on      = [seqera_compute_env.this]
  create_duration = var.wait_duration
}

output "compute_env_id" {
  value = seqera_compute_env.this.compute_env_id
}

output "compute_env_ready" {
  description = "Dependency anchor - use this in depends_on"
  value       = time_sleep.wait.id
}

# Root configuration
module "compute_env" {
  source = "./modules/compute-env"
  # ...
}

resource "seqera_pipeline" "pipeline" {
  depends_on = [module.compute_env.compute_env_ready]
  # ...
}
```

## Known Limitations

- The `time_sleep` approach uses a fixed duration and doesn't actually verify the compute environment status
- If provisioning takes longer than expected, dependent resources may still fail
- If provisioning completes faster, you're waiting unnecessarily

## Related Resources

- [time_sleep resource documentation](https://registry.terraform.io/providers/hashicorp/time/latest/docs/resources/sleep)
- [Terraform depends_on meta-argument](https://www.terraform.io/language/meta-arguments/depends_on)
