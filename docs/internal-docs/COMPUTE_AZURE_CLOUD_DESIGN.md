# Azure Cloud Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_azure_cloud` resource, which manages Azure Cloud (VM-based) compute environments in Seqera Platform.

Azure Cloud compute environments provide compute capacity using direct Azure Virtual Machines managed by Seqera Platform, offering a simpler alternative to Azure Batch for certain workloads.

## Key Differences: Azure Cloud vs Azure Batch

| Feature | Azure Cloud | Azure Batch |
|---------|-------------|-------------|
| **Compute Service** | Direct Azure VMs | Azure Batch service |
| **Management** | Seqera manages VMs | Azure Batch manages jobs |
| **Configuration** | Simpler setup | More configuration options |
| **Scaling** | Basic VM provisioning | Advanced pool management with autoscaling |
| **Control** | Direct VM control | Batch job abstraction |
| **Use Case** | Simpler workflows, quick setup | Complex batch processing, autoscaling needs |

## Resource Structure

```hcl
resource "seqera_compute_azure_cloud" "example" {
  name         = "azure-cloud-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"

  work_directory = "az://mystorageaccount/work"

  # Resource labels
  resource_labels = [
    {
      name  = "environment"
      value = "production"
    },
    {
      name  = "managed_by"
      value = "terraform"
    }
  ]

  # Staging options
  staging_options {
    pre_run_script  = "#!/bin/bash\necho 'Starting workflow'"
    post_run_script = "#!/bin/bash\necho 'Workflow complete'"
  }

  # Nextflow configuration
  nextflow_config = <<-EOF
    process {
      executor = 'azurebatch'
      queue = 'default'
    }
  EOF

  # Environment variables
  environment_variables = {
    "MY_VAR"       = "value"
    "AZURE_REGION" = "eastus"
  }

  # Advanced options
  advanced {
    azure_subscription_id = "00000000-0000-0000-0000-000000000000"
    instance_type         = "Standard_D4s_v3"
  }
}
```

## Field Definitions

### Required Fields

#### `name`
- **Type**: String
- **Required**: Yes
- **Max Length**: 100 characters
- **Description**: Display name for the compute environment
- **Constraints**: Use only alphanumeric characters, dashes, and underscores
- **Example**: `"azure-cloud-prod"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: Azure credentials ID to use for accessing Azure services
- **Reference**: Must reference a valid `seqera_azure_credential` resource
- **Example**: `seqera_azure_credential.main.credentials_id`
- **Notes**: Credentials must have permissions to:
  - Create and manage Azure VMs
  - Access Azure Blob Storage
  - Manage networking and resource groups

#### `region`
- **Type**: String
- **Required**: Yes
- **Description**: Azure location where the workload will be deployed
- **Validation**: Must be a valid Azure region/location identifier
- **Examples**: `"eastus"`, `"westeurope"`, `"southeastasia"`, `"australiaeast"`
- **Notes**: Use Azure region names (not display names)

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: Azure Blob Storage container path for Nextflow work directory
- **Format**: `az://storage-account-name/container-name` or `az://storage-account-name/container-name/path`
- **Example**: `"az://mystorageaccount/nextflow-work"`
- **Constraints**:
  - Storage account must exist
  - Container should exist or be auto-created
  - Credentials must have read/write access
  - Must use `az://` protocol prefix (also referred to as Blob Storage container format)

### Optional Fields

#### `resource_labels`
- **Type**: List of Objects
- **Optional**: Yes
- **Description**: Key-value pairs (name/value tags) associated with resources created by this compute environment
- **Object Structure**:
  - `name` (String): Label/tag key
  - `value` (String): Label/tag value
- **Character Limit**: Combined length should be reasonable for Azure tags
- **Example**:
  ```hcl
  resource_labels = [
    {
      name  = "seqera_workspace"
      value = "test-workspace"
    },
    {
      name  = "environment"
      value = "production"
    },
    {
      name  = "cost_center"
      value = "research"
    },
    {
      name  = "managed_by"
      value = "terraform"
    }
  ]
  ```
- **Notes**:
  - Labels applied as Azure tags to VMs and related resources
  - Useful for cost tracking and resource management
  - Only resource labels with the same name can be used (API constraint)
  - Common default labels: `seqera_workspace`, `Teamstest`

### Staging Options Block

#### `staging_options`
Configuration for workflow staging and lifecycle scripts.

##### `staging_options.pre_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script executed before pipeline launch in the same environment where Nextflow runs
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Setting up environment..."
  export AZURE_ENV=production

  # Download reference data
  az storage blob download-batch \
    --account-name mystorageaccount \
    --source reference-data \
    --destination /mnt/reference
  ```
- **Use Cases**:
  - Load environment modules
  - Download reference data
  - Configure Azure resources
  - Set up directories
  - Validate prerequisites
  - Install additional tools

##### `staging_options.post_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script executed immediately after pipeline completion in the same environment where Nextflow runs
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Pipeline completed with exit code: $NXF_EXIT_STATUS"

  # Archive results
  az storage blob upload-batch \
    --account-name archiveaccount \
    --destination "results/$(date +%Y-%m-%d)" \
    --source /tmp/results

  # Send notification
  echo "Workflow completed" | mail -s "Pipeline Done" user@example.com
  ```
- **Use Cases**:
  - Cleanup temporary files
  - Archive results
  - Send notifications
  - Generate reports
  - Update databases
  - Copy results to different storage

##### `staging_options.nextflow_config`
- **Type**: String
- **Optional**: Yes (field name is `nextflow_config` at root level, not nested in staging_options)
- **Description**: Global Nextflow configuration settings for all pipelines launched with this compute environment
- **Format**: Nextflow configuration DSL
- **Character Limit**: 0/3200 characters
- **Example**:
  ```groovy
  process {
    executor = 'azurebatch'
    queue = 'default'

    errorStrategy = 'retry'
    maxRetries = 3

    cpus = 2
    memory = '4 GB'

    withLabel: big_mem {
      memory = '32 GB'
      cpus = 8
    }

    withLabel: gpu {
      containerOptions = '--gpus all'
    }
  }

  azure {
    batch {
      allowPoolCreation = true
      autoPoolMode = true
    }
    storage {
      accountName = 'mystorageaccount'
    }
  }

  docker {
    enabled = true
    runOptions = '-u $(id -u):$(id -g)'
  }
  ```
- **Use Cases**:
  - Override default executor settings
  - Configure resource requirements
  - Set error handling strategies
  - Define process labels and selectors
  - Configure Docker/Singularity settings
  - Set Azure-specific options

### Environment Variables

#### `environment_variables`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables to set in all compute jobs
- **Example**:
  ```hcl
  environment_variables = {
    "JAVA_OPTS"               = "-Xmx4g"
    "NXF_ANSI_LOG"            = "false"
    "NXF_OPTS"                = "-Xms1g -Xmx4g"
    "AZURE_SUBSCRIPTION_ID"   = "00000000-0000-0000-0000-000000000000"
    "AZURE_REGION"            = "eastus"
    "TMPDIR"                  = "/mnt/batch/tasks/shared"
  }
  ```
- **Notes**:
  - Variables available to all processes in the workflow
  - Useful for configuring tools and runtime behavior
  - Can override default Nextflow settings

### Advanced Options Block

#### `advanced`
Advanced configuration options for fine-tuning the compute environment.

##### `advanced.azure_subscription_id`
- **Type**: String
- **Optional**: Yes
- **Description**: The Azure subscription ID under which resources will be provisioned and billed
- **Format**: UUID (e.g., `"00000000-0000-0000-0000-000000000000"`)
- **Character Limit**: 0/200 characters
- **Example**: `"12345678-1234-1234-1234-123456789012"`
- **Default**: If unspecified, the default subscription from the credentials will be used
- **Notes**:
  - Must be a valid Azure subscription ID
  - Credentials must have access to this subscription
  - Useful when working with multiple subscriptions
  - Resources and costs are billed to this subscription

##### `advanced.instance_type`
- **Type**: String
- **Optional**: Yes
- **Description**: Azure VM size/type for compute nodes
- **Format**: Must be a valid Azure Batch VM type
- **Default**: `"Standard_D32s_v3"` (shown in UI as default)
- **Examples**:
  - General Purpose: `"Standard_D2s_v3"`, `"Standard_D4s_v3"`, `"Standard_D8s_v3"`, `"Standard_D16s_v3"`, `"Standard_D32s_v3"`
  - Compute Optimized: `"Standard_F4s_v2"`, `"Standard_F8s_v2"`, `"Standard_F16s_v2"`, `"Standard_F32s_v2"`
  - Memory Optimized: `"Standard_E4s_v3"`, `"Standard_E8s_v3"`, `"Standard_E16s_v3"`, `"Standard_E32s_v3"`
  - GPU: `"Standard_NC6"`, `"Standard_NC12"`, `"Standard_NC24"`, `"Standard_NV6"`, `"Standard_NV12"`
- **Notes**:
  - Must be available in the specified region
  - Consider CPU, memory, network, and cost requirements
  - GPU instances for deep learning and compute-intensive workloads
  - Larger instances for memory-intensive workflows

## Read-Only Computed Fields

### `compute_env_id`
- **Type**: String
- **Computed**: Yes
- **Description**: Unique identifier for the compute environment assigned by Seqera Platform
- **Example**: `"1a2b3c4d5e6f7g8h"`

### `status`
- **Type**: String
- **Computed**: Yes
- **Description**: Current status of the compute environment
- **Possible Values**: `"CREATING"`, `"AVAILABLE"`, `"INVALID"`, `"DELETING"`

### `message`
- **Type**: String
- **Computed**: Yes
- **Description**: Status message or error details
- **Example**: `"Compute environment is available"`

### `date_created`
- **Type**: String (RFC3339 timestamp)
- **Computed**: Yes
- **Description**: Timestamp when the compute environment was created

### `last_updated`
- **Type**: String (RFC3339 timestamp)
- **Computed**: Yes
- **Description**: Timestamp when the compute environment was last modified

## Implementation Notes

### Validation Rules

1. **region**: Must be a valid Azure region/location identifier
2. **work_directory**: Must start with `az://` and follow format `az://storage-account/container`
3. **azure_subscription_id**: Must be valid UUID format if specified
4. **instance_type**: Must be a valid Azure VM size name
5. **resource_labels**: Each label must have both `name` and `value` fields

### Lifecycle Considerations

- **Create**: Provisions Azure Cloud compute environment configuration in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields (some fields may require replacement)
- **Delete**: Removes compute environment from Seqera Platform (terminates any running VMs)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `region`
- `credentials_id`
- `advanced.azure_subscription_id`

### Mutable Fields

These fields can be updated without replacement:
- `work_directory`
- `resource_labels`
- `staging_options` (pre_run_script, post_run_script)
- `nextflow_config`
- `environment_variables`
- `advanced.instance_type`

### Sensitive Fields

- The referenced `credentials_id` points to sensitive Azure credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration

```hcl
resource "seqera_compute_azure_cloud" "minimal" {
  name           = "azure-cloud-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://mystorageaccount/work"
}
```

### Standard Configuration with Resource Labels

```hcl
resource "seqera_compute_azure_cloud" "standard" {
  name           = "azure-cloud-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "westeurope"
  work_directory = "az://workstorageaccount/nextflow-work"

  resource_labels = [
    {
      name  = "seqera_workspace"
      value = "genomics-workspace"
    },
    {
      name  = "environment"
      value = "production"
    },
    {
      name  = "cost_center"
      value = "research"
    }
  ]

  advanced {
    instance_type = "Standard_D8s_v3"
  }
}
```

### Production Configuration with All Options

```hcl
resource "seqera_compute_azure_cloud" "production" {
  name           = "azure-cloud-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://prodstorageaccount/workflows"

  resource_labels = [
    {
      name  = "seqera_workspace"
      value = "production-workspace"
    },
    {
      name  = "environment"
      value = "production"
    },
    {
      name  = "cost_center"
      value = "genomics"
    },
    {
      name  = "managed_by"
      value = "terraform"
    },
    {
      name  = "compliance"
      value = "hipaa"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Initializing production environment..."
      export AZURE_ENV=production

      # Load required modules
      module load nextflow/23.10.0
      module load java/11

      # Download reference data from blob storage
      az storage blob download-batch \
        --account-name prodstorageaccount \
        --source reference-data \
        --destination /mnt/reference \
        --pattern "*.fa"

      # Validate environment
      echo "Validating prerequisites..."
      nextflow -version
      java -version
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed with exit status: $NXF_EXIT_STATUS"

      # Archive results with timestamp
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      ARCHIVE_PATH="results/run-$TIMESTAMP"

      echo "Archiving results to $ARCHIVE_PATH..."
      az storage blob upload-batch \
        --account-name archiveaccount \
        --destination "$ARCHIVE_PATH" \
        --source /tmp/results \
        --overwrite

      # Generate summary report
      echo "Generating summary report..."
      cat > /tmp/report.txt <<REPORT
      Pipeline Execution Summary
      ==========================
      Timestamp: $TIMESTAMP
      Exit Status: $NXF_EXIT_STATUS
      Archive Location: $ARCHIVE_PATH
REPORT

      az storage blob upload \
        --account-name archiveaccount \
        --container-name reports \
        --name "report-$TIMESTAMP.txt" \
        --file /tmp/report.txt

      # Cleanup
      echo "Cleaning up temporary files..."
      rm -rf /tmp/work/*
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'azurebatch'
      errorStrategy = 'retry'
      maxRetries = 2

      cpus = 4
      memory = '16 GB'
      time = '8h'

      withLabel: small_task {
        cpus = 2
        memory = '8 GB'
      }

      withLabel: intensive {
        cpus = 16
        memory = '64 GB'
        time = '24h'
      }

      withLabel: high_mem {
        memory = '128 GB'
      }
    }

    azure {
      batch {
        allowPoolCreation = true
        autoPoolMode = true
        deletePoolsOnCompletion = true
      }
      storage {
        accountName = 'prodstorageaccount'
      }
    }

    docker {
      enabled = true
      runOptions = '-u $(id -u):$(id -g)'
    }

    report {
      enabled = true
      file = 'pipeline_report.html'
    }

    trace {
      enabled = true
      file = 'pipeline_trace.txt'
    }
  EOF

  environment_variables = {
    "NXF_ANSI_LOG"          = "false"
    "NXF_OPTS"              = "-Xms2g -Xmx8g"
    "AZURE_SUBSCRIPTION_ID" = var.azure_subscription_id
    "AZURE_REGION"          = "eastus"
    "TMPDIR"                = "/mnt/batch/tasks/shared"
    "JAVA_HOME"             = "/usr/lib/jvm/java-11"
  }

  advanced {
    azure_subscription_id = var.azure_subscription_id
    instance_type         = "Standard_D16s_v3"
  }
}
```

### High-Performance Compute Configuration

```hcl
resource "seqera_compute_azure_cloud" "high_performance" {
  name           = "azure-cloud-hpc"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "southcentralus"
  work_directory = "az://hpcstorageaccount/work"

  resource_labels = [
    {
      name  = "workload_type"
      value = "high_performance"
    },
    {
      name  = "priority"
      value = "critical"
    }
  ]

  advanced {
    instance_type = "Standard_F32s_v2"  # Compute optimized
  }

  nextflow_config = <<-EOF
    process {
      cpus = 32
      memory = '64 GB'

      withLabel: parallel {
        maxForks = 16
      }
    }
  EOF

  environment_variables = {
    "NXF_OPTS" = "-Xms4g -Xmx16g"
    "OMP_NUM_THREADS" = "32"
  }
}
```

### GPU-Enabled Configuration

```hcl
resource "seqera_compute_azure_cloud" "gpu" {
  name           = "azure-cloud-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://gpustorageaccount/work"

  resource_labels = [
    {
      name  = "compute_type"
      value = "gpu"
    },
    {
      name  = "ml_workload"
      value = "true"
    }
  ]

  advanced {
    instance_type = "Standard_NC6"  # GPU instance
  }

  nextflow_config = <<-EOF
    process {
      withLabel: gpu {
        containerOptions = '--gpus all'
        accelerator = 1
      }
    }
  EOF

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0"
    "NVIDIA_VISIBLE_DEVICES" = "all"
  }
}
```

### Memory-Optimized Configuration

```hcl
resource "seqera_compute_azure_cloud" "memory_optimized" {
  name           = "azure-cloud-highmem"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "westus2"
  work_directory = "az://memstorageaccount/work"

  resource_labels = [
    {
      name  = "workload_type"
      value = "memory_intensive"
    }
  ]

  advanced {
    instance_type = "Standard_E32s_v3"  # Memory optimized
  }

  nextflow_config = <<-EOF
    process {
      memory = '128 GB'

      withLabel: extreme_mem {
        memory = '256 GB'
      }
    }
  EOF

  environment_variables = {
    "NXF_OPTS" = "-Xms8g -Xmx32g"
  }
}
```

### Development/Testing Configuration

```hcl
resource "seqera_compute_azure_cloud" "dev" {
  name           = "azure-cloud-dev"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus2"
  work_directory = "az://devstorageaccount/work"

  resource_labels = [
    {
      name  = "environment"
      value = "development"
    },
    {
      name  = "owner"
      value = "dev-team"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Development environment setup"
      export DEBUG=true
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Development run completed"
      # Keep logs for debugging
    EOF
  }

  advanced {
    instance_type = "Standard_D4s_v3"  # Smaller, cost-effective
  }

  environment_variables = {
    "NXF_DEBUG" = "2"
    "ENV"       = "development"
  }
}
```

### Multi-Region Configuration Pattern

```hcl
# Variables for multi-region deployment
variable "regions" {
  type = map(object({
    storage_account = string
    instance_type   = string
  }))
  default = {
    eastus = {
      storage_account = "eastusstorage"
      instance_type   = "Standard_D8s_v3"
    }
    westeurope = {
      storage_account = "westeuropestorage"
      instance_type   = "Standard_D8s_v3"
    }
  }
}

# Create compute environments in multiple regions
resource "seqera_compute_azure_cloud" "multi_region" {
  for_each = var.regions

  name           = "azure-cloud-${each.key}"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = each.key
  work_directory = "az://${each.value.storage_account}/work"

  resource_labels = [
    {
      name  = "region"
      value = each.key
    },
    {
      name  = "deployment"
      value = "multi-region"
    }
  ]

  advanced {
    instance_type = each.value.instance_type
  }

  environment_variables = {
    "AZURE_REGION" = each.key
  }
}
```

## API Mapping

### Seqera Platform API Endpoints

- **Create**: `POST /compute-envs`
- **Read**: `GET /compute-envs/{computeEnvId}`
- **Update**: `PUT /compute-envs/{computeEnvId}`
- **Delete**: `DELETE /compute-envs/{computeEnvId}`
- **List**: `GET /compute-envs?workspaceId={workspaceId}`

### Request/Response Schema Mapping

The Terraform resource fields map to the Seqera Platform API as follows:

- Resource `name` → API `config.name`
- Resource `region` → API `config.region`
- Resource `work_directory` → API `config.workDir`
- Resource `resource_labels` → API `config.resourceLabels`
- Resource `staging_options.pre_run_script` → API `config.preRunScript`
- Resource `staging_options.post_run_script` → API `config.postRunScript`
- Resource `nextflow_config` → API `config.nextflowConfig`
- Resource `environment_variables` → API `config.environment`
- Resource `advanced.azure_subscription_id` → API `config.azureSubscriptionId`
- Resource `advanced.instance_type` → API `config.instanceType`

### Platform Type

- **API Value**: `"azure-cloud"` or `"azure-batch"` (with Cloud-specific configuration)
- **Config Type**: `"AzureBatchComputeConfig"` (reuses Azure Batch config type but with Cloud mode)

## Related Resources

- `seqera_azure_credential` - Azure credentials used by the compute environment
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Comparison: Azure Cloud vs Azure Batch

| Aspect | Azure Cloud | Azure Batch |
|--------|-------------|-------------|
| **Setup Complexity** | Simpler | More complex |
| **Configuration** | Minimal options | Extensive options (forge/manual modes) |
| **VM Management** | Direct VM provisioning | Batch pool management |
| **Autoscaling** | Basic | Advanced with pool autoscaling |
| **Features** | Core features | Wave, Fusion v2, container registry |
| **Control Level** | Moderate | High |
| **Cost** | Direct VM costs | Batch service overhead |
| **Best For** | Simple workflows, quick setup | Complex workflows, autoscaling, advanced features |

## Azure Subscription Requirements

### Required Permissions

The Azure credentials must have the following permissions:

#### Virtual Machine Permissions
- `Microsoft.Compute/virtualMachines/read`
- `Microsoft.Compute/virtualMachines/write`
- `Microsoft.Compute/virtualMachines/delete`
- `Microsoft.Compute/disks/*`

#### Storage Account Permissions
- `Microsoft.Storage/storageAccounts/read`
- `Microsoft.Storage/storageAccounts/listKeys/action`
- `Microsoft.Storage/storageAccounts/blobServices/containers/*`

#### Networking Permissions
- `Microsoft.Network/virtualNetworks/read`
- `Microsoft.Network/networkInterfaces/*`
- `Microsoft.Network/publicIPAddresses/*`
- `Microsoft.Network/networkSecurityGroups/*`

#### Resource Group Permissions
- `Microsoft.Resources/subscriptions/resourceGroups/read`
- `Microsoft.Resources/subscriptions/resourceGroups/write`

### Azure Subscription Configuration

- **Subscription**: Must have sufficient quota for desired VM types
- **Resource Quotas**: Check regional VM core quotas
- **Storage Account**: Must exist with appropriate access configuration
- **Networking**: VNet and subnet configuration if required

## Migration Guide

### Migrating from Azure Batch to Azure Cloud

If you're currently using Azure Batch and want to simplify to Azure Cloud:

```hcl
# Before: Azure Batch with Forge
resource "seqera_compute_azure_batch" "old" {
  name           = "azure-compute"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://mystorageaccount/work"

  forge {
    vm_type  = "Standard_D4s_v3"
    vm_count = 1
  }
}

# After: Azure Cloud (simpler)
resource "seqera_compute_azure_cloud" "new" {
  name           = "azure-compute"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://mystorageaccount/work"

  advanced {
    instance_type = "Standard_D4s_v3"
  }
}
```

**Key Changes:**
- Remove `forge` or `manual` blocks
- Move `vm_type` to `advanced.instance_type`
- Remove Batch-specific features (autoscale, managed_identity, etc.)
- Simplified configuration overall

### When to Choose Azure Cloud

Choose Azure Cloud when:
- You want simpler configuration
- You don't need advanced autoscaling
- You don't need Wave or Fusion v2 features
- You have predictable workloads
- You prefer direct VM control

Choose Azure Batch when:
- You need autoscaling capabilities
- You want Wave containers or Fusion v2
- You have variable workloads
- You need advanced pool management
- You want managed identity integration

## References

- [Azure Virtual Machines Documentation](https://docs.microsoft.com/en-us/azure/virtual-machines/)
- [Azure Blob Storage Documentation](https://docs.microsoft.com/en-us/azure/storage/blobs/)
- [Nextflow Azure Documentation](https://www.nextflow.io/docs/latest/azure.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Azure VM Sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes)
