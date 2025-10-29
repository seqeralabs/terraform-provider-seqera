# Azure Batch Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_azure_batch` resource, which manages Azure Batch compute environments in Seqera Platform.

Azure Batch compute environments provide scalable compute capacity for running Nextflow workflows on Azure using the Azure Batch service.

## Resource Structure

```hcl
resource "seqera_compute_azure_batch" "example" {
  name         = "azure-batch-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"

  work_directory = "az://mystorageaccount/work"

  # Batch Forge configuration (managed pool)
  forge {
    vm_type                    = "Standard_D4s_v3"
    vm_count                   = 1
    autoscale                  = true
    managed_identity_client_id = "00000000-0000-0000-0000-000000000000"
    dispose_resources          = false
  }

  # OR Manual configuration (existing pool)
  # manual {
  #   pool_id         = "my-existing-pool"
  #   batch_account   = "mybatchaccount"
  #   resource_group  = "my-resource-group"
  # }

  # Wave containers
  wave {
    enabled          = true
    strategy         = "conda,container"
    freeze_mode      = true
    build_repository = "myregistry.azurecr.io/wave/builds"
    cache_repository = "myregistry.azurecr.io/wave/cache"
  }

  # Fusion v2
  fusion_v2 {
    enabled           = true
    fusion_log_level  = "DEBUG"
    fusion_log_output = "/var/log/fusion.log"
    tags_pattern      = ".*"
  }

  # Container registry credentials
  container_registry_credentials = [
    seqera_container_registry_credential.acr.credentials_id
  ]

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
    "MY_VAR"        = "value"
    "AZURE_REGION"  = "eastus"
  }

  # Advanced options
  advanced {
    jobs_cleanup_policy = "on_success"
    token_duration      = "10h"
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
- **Example**: `"azure-batch-prod"`

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
- **Notes**: Credentials must include both Azure Batch and Storage account details

#### `region`
- **Type**: String
- **Required**: Yes
- **Description**: Azure region where the Batch compute environment will be created
- **Validation**: Must be a valid Azure region identifier
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
  - Container must exist or be auto-created
  - Credentials must have read/write access
  - Must use `az://` protocol prefix

### Configuration Mode (Mutually Exclusive)

You must specify exactly one of `forge` or `manual` blocks.

### Forge Configuration Block (Managed Pool)

#### `forge`
Configuration for Seqera-managed Azure Batch pool using Batch Forge.

##### `forge.vm_type`
- **Type**: String
- **Required**: Yes (when using forge)
- **Description**: Azure VM size/type for compute nodes
- **Examples**:
  - General Purpose: `"Standard_D2s_v3"`, `"Standard_D4s_v3"`, `"Standard_D8s_v3"`
  - Compute Optimized: `"Standard_F4s_v2"`, `"Standard_F8s_v2"`, `"Standard_F16s_v2"`
  - Memory Optimized: `"Standard_E4s_v3"`, `"Standard_E8s_v3"`, `"Standard_E16s_v3"`
  - GPU: `"Standard_NC6"`, `"Standard_NC12"`, `"Standard_NV6"`
- **Default**: `"Standard_D4s_v3"` (if not specified, shown in UI)
- **Notes**:
  - Must be available in the specified region
  - Consider CPU, memory, and cost requirements
  - GPU VMs for deep learning workloads

##### `forge.vm_count`
- **Type**: Integer
- **Required**: Yes (when using forge)
- **Description**: Number of VMs in the pool
- **Default**: `1`
- **Range**: 1 to service limits
- **Example**: `4`
- **Notes**:
  - When autoscale is enabled, represents maximum pool size
  - When autoscale is disabled, represents fixed pool size
  - Subject to Azure Batch account quotas

##### `forge.autoscale`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable pool autoscaling to automatically adjust pool size based on workload
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - When enabled, pool automatically scales up and down
  - Helps optimize costs by scaling to zero when idle
  - Pool scales up when jobs are queued

##### `forge.managed_identity_client_id`
- **Type**: String
- **Optional**: Yes
- **Description**: Client ID of a managed identity attached to the Azure Batch Pool
- **Format**: UUID (e.g., `"00000000-0000-0000-0000-000000000000"`)
- **Example**: `"12345678-1234-1234-1234-123456789012"`
- **Notes**:
  - Enables nodes to access Azure resources without explicit credentials
  - Managed identity must be associated with the batch account
  - Useful for accessing Azure Storage, Key Vault, etc.

##### `forge.dispose_resources`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Whether to preserve the Batch pool after compute environment deletion
- **Default**: `true` (pool is deleted with compute environment)
- **Example**: `false`
- **Notes**:
  - `true` - Pool deleted when compute environment is deleted
  - `false` - Pool preserved independently from compute environment lifecycle
  - Useful for reusing pools across multiple compute environments

### Manual Configuration Block (Existing Pool)

#### `manual`
Configuration for using an existing Azure Batch pool.

##### `manual.pool_id`
- **Type**: String
- **Required**: Yes (when using manual)
- **Description**: ID of the existing Azure Batch pool
- **Example**: `"my-existing-pool"`
- **Notes**: Pool must already exist in the specified Batch account

##### `manual.batch_account`
- **Type**: String
- **Required**: Yes (when using manual)
- **Description**: Name of the Azure Batch account containing the pool
- **Example**: `"mybatchaccount"`
- **Notes**: Must match the batch account in credentials

##### `manual.resource_group`
- **Type**: String
- **Required**: Yes (when using manual)
- **Description**: Azure resource group name containing the Batch account
- **Example**: `"my-resource-group"`
- **Notes**: Resource group must exist in the subscription

### Optional Feature Blocks

#### `wave`
Wave containers configuration for private container repository access and preprocessing.

##### `wave.enabled`
- **Type**: Boolean
- **Required**: Yes (when wave block is present)
- **Description**: Enable Wave containers service
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Wave allows access to private container repositories
  - Enables on-demand container building
  - Supports Conda package installations

##### `wave.strategy`
- **Type**: String
- **Optional**: Yes
- **Description**: Build strategy for Wave containers
- **Format**: Comma-separated values
- **Examples**: `"conda"`, `"container"`, `"conda,container"`
- **Notes**: Defines how Wave builds and manages container images

##### `wave.freeze_mode`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Freeze container images to specific versions
- **Default**: `false`
- **Example**: `true`
- **Notes**: Ensures reproducibility by freezing container versions

##### `wave.build_repository`
- **Type**: String
- **Optional**: Yes
- **Description**: Container registry repository for Wave builds
- **Format**: `registry.azurecr.io/repository/path`
- **Example**: `"myregistry.azurecr.io/wave/builds"`
- **Notes**: Repository for storing built container images

##### `wave.cache_repository`
- **Type**: String
- **Optional**: Yes
- **Description**: Container registry repository for Wave cache
- **Format**: `registry.azurecr.io/repository/path`
- **Example**: `"myregistry.azurecr.io/wave/cache"`
- **Notes**: Repository for caching container layers

#### `fusion_v2`
Fusion v2 configuration for optimized Azure Blob Storage access.

##### `fusion_v2.enabled`
- **Type**: Boolean
- **Required**: Yes (when fusion_v2 block is present)
- **Description**: Enable Fusion v2 virtual file system
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Provides virtual file system for efficient Azure Blob Storage access
  - Improves performance with lazy loading
  - Reduces data transfer costs

##### `fusion_v2.fusion_log_level`
- **Type**: String
- **Optional**: Yes
- **Description**: Logging level for Fusion v2
- **Allowed Values**: `"TRACE"`, `"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"`
- **Default**: `"INFO"`
- **Example**: `"DEBUG"`
- **Notes**: Higher verbosity useful for troubleshooting

##### `fusion_v2.fusion_log_output`
- **Type**: String
- **Optional**: Yes
- **Description**: Path to Fusion v2 log file
- **Example**: `"/var/log/fusion.log"`
- **Notes**: Location where Fusion writes its logs

##### `fusion_v2.tags_pattern`
- **Type**: String
- **Optional**: Yes
- **Description**: Regular expression pattern for blob object tags
- **Default**: `".*"` (all tags)
- **Example**: `"prod-.*"`
- **Notes**: Filter which blob tags Fusion processes

### Container Registry Credentials

#### `container_registry_credentials`
- **Type**: List of Strings
- **Optional**: Yes
- **Description**: List of container registry credential IDs for private registries
- **Example**:
  ```hcl
  container_registry_credentials = [
    seqera_container_registry_credential.acr.credentials_id,
    seqera_container_registry_credential.docker.credentials_id
  ]
  ```
- **Notes**:
  - Each credential must reference a valid container registry credential resource
  - Required for pulling images from private registries
  - Supports Azure Container Registry (ACR), Docker Hub, etc.

### Resource Labels

#### `resource_labels`
- **Type**: List of Objects
- **Optional**: Yes
- **Description**: Key-value pairs (tags) applied to Azure resources created by this compute environment
- **Object Structure**:
  - `name` (String): Label/tag key
  - `value` (String): Label/tag value
- **Example**:
  ```hcl
  resource_labels = [
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
  - Labels applied to Batch pool and related resources
  - Useful for cost tracking and resource management
  - Only labels with the same name can be used (API constraint)
  - Common default labels: `seqera_workspace`, `transmitter`

### Staging Options Block

#### `staging_options`
Configuration for workflow staging and lifecycle scripts.

##### `staging_options.pre_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script executed before workflow starts
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Setting up environment..."
  export AZURE_ENV=production
  az storage blob download --account-name mystorageaccount \
    --container-name reference-data --name genome.fa --file /mnt/ref/genome.fa
  ```
- **Use Cases**:
  - Load environment modules
  - Download reference data
  - Set up tools or dependencies
  - Validate prerequisites
  - Configure Azure resources

##### `staging_options.post_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script executed after workflow completes
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Pipeline completed with exit code: $NXF_EXIT_STATUS"
  # Archive results
  az storage blob upload-batch --account-name mystorageaccount \
    --destination results --source /tmp/results
  ```
- **Use Cases**:
  - Cleanup temporary files
  - Archive results
  - Send notifications
  - Generate reports
  - Update databases

### Nextflow Configuration

#### `nextflow_config`
- **Type**: String
- **Optional**: Yes
- **Description**: Custom Nextflow configuration for this compute environment
- **Format**: Nextflow configuration DSL
- **Character Limit**: 0/3200 characters
- **Example**:
  ```groovy
  process {
    executor = 'azurebatch'
    queue = 'my-pool'

    errorStrategy = 'retry'
    maxRetries = 3

    cpus = 2
    memory = '4 GB'

    withLabel: big_mem {
      memory = '32 GB'
    }
  }

  azure {
    batch {
      allowPoolCreation = true
      autoPoolMode = true
      deletePoolsOnCompletion = true
    }
    storage {
      accountName = 'mystorageaccount'
      accountKey = env.AZURE_STORAGE_KEY
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

### Environment Variables

#### `environment_variables`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables set in all compute jobs
- **Example**:
  ```hcl
  environment_variables = {
    "JAVA_OPTS"               = "-Xmx4g"
    "NXF_ANSI_LOG"            = "false"
    "NXF_OPTS"                = "-Xms1g -Xmx4g"
    "AZURE_SUBSCRIPTION_ID"   = "00000000-0000-0000-0000-000000000000"
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

##### `advanced.jobs_cleanup_policy`
- **Type**: String
- **Optional**: Yes
- **Description**: Policy for automatic deletion of Azure Batch jobs
- **Allowed Values**:
  - `"on_success"` - Delete jobs only when pipeline completes successfully
  - `"always"` - Always delete jobs after pipeline completion (success or failure)
  - `"never"` - Never automatically delete jobs
- **Default**: `"on_success"`
- **Example**: `"always"`
- **Notes**:
  - Helps manage Azure Batch job quotas
  - `never` useful for debugging failed pipelines
  - `always` helps with cleanup and cost management

##### `advanced.token_duration`
- **Type**: String
- **Optional**: Yes
- **Description**: Duration of the shared access signature (SAS) token for Azure Storage
- **Format**: Duration string (e.g., `"10h"`, `"30m"`, `"2d"`)
- **Default**: `"10h"`
- **Range**: Minimum 1 minute to maximum 7 days
- **Examples**: `"1h"`, `"6h"`, `"12h"`, `"24h"`, `"2d"`
- **Notes**:
  - SAS token created by Nextflow when `sasToken` option not specified
  - Token must be valid for entire workflow duration
  - Longer workflows need longer token duration
  - Consider security vs convenience tradeoff

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

1. **region**: Must be a valid Azure region identifier
2. **work_directory**: Must start with `az://` and follow format `az://storage-account/container`
3. **Config mode**: Exactly one of `forge` or `manual` must be specified
4. **forge.vm_count**: Must be positive integer
5. **jobs_cleanup_policy**: Must be one of `"on_success"`, `"always"`, `"never"`
6. **token_duration**: Must be valid duration format
7. **managed_identity_client_id**: Must be valid UUID format if specified
8. **resource_labels**: Each label must have both `name` and `value` fields

### Lifecycle Considerations

- **Create**: Provisions Azure Batch compute environment and Seqera Platform configuration
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields (some fields may require replacement)
- **Delete**: Removes compute environment from Seqera Platform
  - If `forge.dispose_resources = true`, deletes the Batch pool
  - If `forge.dispose_resources = false`, preserves the Batch pool

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `region`
- `credentials_id`
- `forge` vs `manual` (switching config modes)
- `manual.pool_id`
- `manual.batch_account`
- `manual.resource_group`

### Mutable Fields

These fields can be updated without replacement:
- `work_directory`
- `forge.vm_type`
- `forge.vm_count`
- `forge.autoscale`
- `forge.managed_identity_client_id`
- `wave` configuration
- `fusion_v2` configuration
- `container_registry_credentials`
- `resource_labels`
- `staging_options`
- `nextflow_config`
- `environment_variables`
- `advanced` options

### Sensitive Fields

- The referenced `credentials_id` points to sensitive Azure credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration with Batch Forge

```hcl
resource "seqera_compute_azure_batch" "minimal" {
  name           = "azure-batch-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://mystorageaccount/work"

  forge {
    vm_type  = "Standard_D4s_v3"
    vm_count = 1
  }
}
```

### Autoscaling Configuration

```hcl
resource "seqera_compute_azure_batch" "autoscale" {
  name           = "azure-batch-autoscale"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://mystorageaccount/work"

  forge {
    vm_type   = "Standard_F8s_v2"
    vm_count  = 10  # Maximum pool size
    autoscale = true
  }

  advanced {
    jobs_cleanup_policy = "on_success"
  }
}
```

### Manual Configuration with Existing Pool

```hcl
resource "seqera_compute_azure_batch" "manual" {
  name           = "azure-batch-manual"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "westeurope"
  work_directory = "az://prodstorageaccount/work"

  manual {
    pool_id        = "existing-prod-pool"
    batch_account  = "prodbatchaccount"
    resource_group = "production-resources"
  }
}
```

### Production with Wave and Fusion

```hcl
resource "seqera_compute_azure_batch" "production" {
  name           = "azure-batch-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://prodstorageaccount/nextflow-work"

  forge {
    vm_type                    = "Standard_D8s_v3"
    vm_count                   = 20
    autoscale                  = true
    managed_identity_client_id = "12345678-1234-1234-1234-123456789012"
    dispose_resources          = false
  }

  wave {
    enabled          = true
    strategy         = "conda,container"
    freeze_mode      = true
    build_repository = "prodregistry.azurecr.io/wave/builds"
    cache_repository = "prodregistry.azurecr.io/wave/cache"
  }

  fusion_v2 {
    enabled          = true
    fusion_log_level = "INFO"
  }

  container_registry_credentials = [
    seqera_container_registry_credential.acr.credentials_id
  ]

  resource_labels = [
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
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Initializing production environment..."
      export AZURE_ENV=production

      # Download reference data
      az storage blob download-batch \
        --account-name prodstorageaccount \
        --source reference-data \
        --destination /mnt/reference
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed"

      # Archive results with timestamp
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      az storage blob upload-batch \
        --account-name archiveaccount \
        --destination "results/$TIMESTAMP" \
        --source /results
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'azurebatch'
      errorStrategy = 'retry'
      maxRetries = 2

      cpus = 4
      memory = '16 GB'

      withLabel: intensive {
        cpus = 16
        memory = '64 GB'
      }
    }

    azure {
      batch {
        allowPoolCreation = false
        autoPoolMode = false
      }
    }
  EOF

  environment_variables = {
    "NXF_ANSI_LOG"          = "false"
    "NXF_OPTS"              = "-Xms2g -Xmx8g"
    "AZURE_SUBSCRIPTION_ID" = var.azure_subscription_id
    "TMPDIR"                = "/mnt/batch/tasks/shared"
  }

  advanced {
    jobs_cleanup_policy = "on_success"
    token_duration      = "24h"
  }
}
```

### GPU-Enabled Compute

```hcl
resource "seqera_compute_azure_batch" "gpu" {
  name           = "azure-batch-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "eastus"
  work_directory = "az://gpustorageaccount/work"

  forge {
    vm_type  = "Standard_NC6"  # GPU VM type
    vm_count = 4
  }

  nextflow_config = <<-EOF
    process {
      withLabel: gpu {
        containerOptions = '--gpus all'
      }
    }
  EOF

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0"
  }
}
```

### Development with Resource Preservation

```hcl
resource "seqera_compute_azure_batch" "dev" {
  name           = "azure-batch-dev"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_azure_credential.main.credentials_id
  region         = "westus2"
  work_directory = "az://devstorageaccount/work"

  forge {
    vm_type           = "Standard_D2s_v3"
    vm_count          = 2
    dispose_resources = false  # Preserve pool for reuse
  }

  advanced {
    jobs_cleanup_policy = "never"  # Keep jobs for debugging
    token_duration      = "6h"
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
- Resource `forge.vm_type` → API `config.forge.vmType`
- Resource `forge.vm_count` → API `config.forge.vmCount`
- Resource `forge.autoscale` → API `config.forge.autoscale`
- Resource `manual.pool_id` → API `config.manual.poolId`
- etc.

### Platform Type

- **API Value**: `"azure-batch"`
- **Config Type**: `"AzureBatchComputeConfig"`

## Related Resources

- `seqera_azure_credential` - Azure credentials used by the compute environment
- `seqera_container_registry_credential` - Container registry credentials for private registries
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Comparison: Forge vs Manual

| Aspect | Forge (Managed) | Manual (Existing Pool) |
|--------|-----------------|------------------------|
| **Setup** | Automatic pool creation | Use existing pool |
| **Management** | Seqera manages pool | You manage pool |
| **Flexibility** | Auto-configuration | Full control |
| **Autoscaling** | Built-in support | Depends on pool config |
| **Lifecycle** | Tied to compute env (unless preserved) | Independent |
| **Best For** | Quick start, standard workflows | Custom requirements, shared pools |

## Azure Batch Account Requirements

### Required Permissions

The Azure credentials must have the following permissions:

#### Batch Account Permissions
- `Microsoft.Batch/batchAccounts/read`
- `Microsoft.Batch/batchAccounts/pools/read`
- `Microsoft.Batch/batchAccounts/pools/write` (for Forge mode)
- `Microsoft.Batch/batchAccounts/pools/delete` (for Forge mode with dispose_resources)
- `Microsoft.Batch/batchAccounts/jobs/*`

#### Storage Account Permissions
- `Microsoft.Storage/storageAccounts/read`
- `Microsoft.Storage/storageAccounts/listKeys/action`
- `Microsoft.Storage/storageAccounts/blobServices/containers/*`

### Batch Account Configuration

- **Pool Allocation Mode**: Either User Subscription or Batch Service
- **Identity**: Managed identity if using `managed_identity_client_id`
- **Networking**: Appropriate VNet configuration if required
- **Quotas**: Sufficient core quotas for desired VM types

## References

- [Azure Batch Documentation](https://docs.microsoft.com/en-us/azure/batch/)
- [Nextflow Azure Batch Executor](https://www.nextflow.io/docs/latest/azure.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Azure VM Sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes)
