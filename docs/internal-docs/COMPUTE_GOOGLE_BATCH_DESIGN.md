# Google Batch Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_google_batch` resource, which manages Google Batch compute environments in Seqera Platform.

Google Batch compute environments provide scalable compute capacity for running Nextflow workflows on Google Cloud using the Google Cloud Batch service.

## Resource Structure

```hcl
resource "seqera_compute_google_batch" "example" {
  name         = "google-batch-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"

  work_directory = "gs://my-bucket/work"

  # Spot VMs
  spot = true

  # Wave containers
  wave {
    enabled          = true
    strategy         = "conda,container"
    freeze_mode      = true
    build_repository = "gcr.io/my-project/wave/builds"
    cache_repository = "gcr.io/my-project/wave/cache"
  }

  # Fusion v2 (requires Wave)
  fusion_v2 {
    enabled           = true
    fusion_log_level  = "DEBUG"
    fusion_log_output = "/var/log/fusion.log"
    tags_pattern      = ".*"
  }

  # Resource labels
  resource_labels = [
    {
      name  = "seqera_workspace"
      value = "test-workspace"
    },
    {
      name  = "environment"
      value = "production"
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
      executor = 'google-batch'
    }
  EOF

  # Environment variables
  environment_variables = {
    "MY_VAR"     = "value"
    "GCP_REGION" = "us-central1"
  }

  # Advanced options
  advanced {
    use_private_address          = false
    boot_disk_size               = 50
    head_job_cpus                = 2
    head_job_memory              = 2048
    service_account_email        = "my-service-account@my-project.iam.gserviceaccount.com"
    vpc                          = "my-vpc-network"
    subnet                       = "my-subnet"
    head_job_instance_template   = "projects/my-project/global/instanceTemplates/head-template"
    compute_jobs_instance_template = "projects/my-project/global/instanceTemplates/compute-template"
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
- **Example**: `"google-batch-prod"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: Google Cloud credentials ID to use for accessing Google Cloud services
- **Reference**: Must reference a valid `seqera_google_credential` resource
- **Example**: `seqera_google_credential.main.credentials_id`
- **Notes**: Credentials must have permissions to:
  - Use Google Cloud Batch API
  - Access Google Cloud Storage
  - Manage Compute Engine instances

#### `location`
- **Type**: String
- **Required**: Yes
- **Description**: The Google Cloud location where the job executions are deployed to Google Batch API
- **Validation**: Must be a valid Google Cloud region or zone
- **Examples**: `"us-central1"`, `"europe-west1"`, `"asia-southeast1"`, `"us-east1"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Use region names (e.g., `us-central1`) for regional resources
  - Work directory bucket should be in the same region
  - Some features may be region-specific

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: Google Storage bucket path for Nextflow work directory
- **Format**: `gs://bucket-name/path`
- **Example**: `"gs://my-nextflow-bucket/work"`
- **Character Limit**: 0/400 characters
- **Constraints**:
  - Must start with `gs://`
  - Bucket must exist
  - Credentials must have read/write access
  - Should be in the same region as the compute environment for best performance

### Optional Fields

#### `spot`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Use Spot virtual machines (preemptible VMs)
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Spot VMs are significantly cheaper than regular VMs
  - Can be preempted (interrupted) with 30-second notice
  - Good for fault-tolerant workloads
  - Recommended for cost optimization with proper retry strategies

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
  - Required for Fusion v2

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
- **Format**: `gcr.io/project-id/repository/path` or `us-docker.pkg.dev/project-id/repository`
- **Examples**:
  - `"gcr.io/my-project/wave/builds"`
  - `"us-docker.pkg.dev/my-project/wave-repo/builds"`
- **Notes**: Repository for storing built container images (GCR or Artifact Registry)

##### `wave.cache_repository`
- **Type**: String
- **Optional**: Yes
- **Description**: Container registry repository for Wave cache
- **Format**: `gcr.io/project-id/repository/path` or `us-docker.pkg.dev/project-id/repository`
- **Examples**:
  - `"gcr.io/my-project/wave/cache"`
  - `"us-docker.pkg.dev/my-project/wave-repo/cache"`
- **Notes**: Repository for caching container layers

#### `fusion_v2`
Fusion v2 configuration for optimized Google Cloud Storage access.

##### `fusion_v2.enabled`
- **Type**: Boolean
- **Required**: Yes (when fusion_v2 block is present)
- **Description**: Enable Fusion v2 virtual distributed file system
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Provides virtual file system for efficient Google Cloud Storage access
  - Improves performance with lazy loading
  - Reduces data transfer costs
  - **Requires Wave containers service to be enabled**

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
- **Description**: Regular expression pattern for GCS object labels/tags
- **Default**: `".*"` (all tags)
- **Example**: `"prod-.*"`
- **Notes**: Filter which object labels Fusion processes

### Resource Labels

#### `resource_labels`
- **Type**: List of Objects
- **Optional**: Yes
- **Description**: Key-value pairs (labels) associated with resources created by this compute environment
- **Object Structure**:
  - `name` (String): Label key
  - `value` (String): Label value
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
      value = "genomics"
    },
    {
      name  = "managed_by"
      value = "terraform"
    }
  ]
  ```
- **Notes**:
  - Labels applied to Batch jobs and VM instances
  - Useful for cost tracking and resource management
  - Only one resource label with the same name can be used (API constraint)
  - Default resource labels are pre-filled (e.g., `seqera_workspace`, `Teamstest`)

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
  export GCP_ENV=production

  # Download reference data from GCS
  gsutil -m cp -r gs://reference-data-bucket/genome /mnt/reference/

  # Install additional tools
  apt-get update && apt-get install -y jq
  ```
- **Use Cases**:
  - Load environment modules
  - Download reference data from GCS
  - Configure GCP resources
  - Set up directories
  - Install additional dependencies
  - Validate prerequisites

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

  # Archive results to GCS
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  gsutil -m cp -r /tmp/results gs://archive-bucket/results-$TIMESTAMP/

  # Send notification via Pub/Sub
  gcloud pubsub topics publish pipeline-complete \
    --message "Pipeline completed at $TIMESTAMP with status $NXF_EXIT_STATUS"
  ```
- **Use Cases**:
  - Cleanup temporary files
  - Archive results to GCS
  - Send notifications via Pub/Sub
  - Generate reports
  - Update BigQuery or other databases
  - Trigger downstream workflows

### Nextflow Configuration

#### `nextflow_config`
- **Type**: String
- **Optional**: Yes
- **Description**: Global Nextflow configuration settings for all pipelines launched with this compute environment
- **Format**: Nextflow configuration DSL
- **Character Limit**: 0/3200 characters
- **Example**:
  ```groovy
  process {
    executor = 'google-batch'

    errorStrategy = 'retry'
    maxRetries = 3

    cpus = 2
    memory = '4 GB'
    disk = '20 GB'

    withLabel: big_mem {
      memory = '32 GB'
      cpus = 8
    }

    withLabel: gpu {
      accelerator = [request: 1, type: 'nvidia-tesla-t4']
    }
  }

  google {
    project = 'my-project-id'
    location = 'us-central1'
    batch {
      spot = true
      bootDiskSize = '50 GB'
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
  ```
- **Use Cases**:
  - Override default executor settings
  - Configure resource requirements
  - Set error handling strategies
  - Define process labels and selectors
  - Configure Docker/Singularity settings
  - Set Google Batch-specific options

### Environment Variables

#### `environment_variables`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables to set in all compute jobs
- **Example**:
  ```hcl
  environment_variables = {
    "JAVA_OPTS"              = "-Xmx4g"
    "NXF_ANSI_LOG"           = "false"
    "NXF_OPTS"               = "-Xms1g -Xmx4g"
    "GOOGLE_CLOUD_PROJECT"   = "my-project-id"
    "GCP_REGION"             = "us-central1"
    "TMPDIR"                 = "/tmp"
  }
  ```
- **Notes**:
  - Variables available to all processes in the workflow
  - Useful for configuring tools and runtime behavior
  - Can override default Nextflow settings

### Advanced Options Block

#### `advanced`
Advanced configuration options for fine-tuning the compute environment.

##### `advanced.use_private_address`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Do not attach a public IP address to the VM
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - When enabled, only Google internal services are accessible
  - VMs cannot access public internet directly
  - Requires VPC with appropriate routing (Cloud NAT or Private Google Access)
  - Improves security by preventing external access
  - Useful for compliance requirements

##### `advanced.boot_disk_size`
- **Type**: Integer
- **Optional**: Yes
- **Description**: Boot disk size in GB
- **Default**: `50` (typical default)
- **Range**: Minimum 10 GB to thousands of GB
- **Example**: `100`
- **Character Limit**: 0/4 characters (for the input field)
- **Notes**:
  - Larger disk for container images and temporary files
  - Consider workflow storage requirements
  - Standard persistent disk used by default

##### `advanced.head_job_cpus`
- **Type**: Integer
- **Optional**: Yes
- **Description**: The number of CPUs allocated for the Nextflow runner job (head job)
- **Default**: `1` or `2` (typical)
- **Range**: 1 to 96 (depends on machine type limits)
- **Example**: `4`
- **Character Limit**: 0/4 characters
- **Notes**:
  - Head job orchestrates the workflow
  - More CPUs may help with large DAG workflows
  - Consider workflow complexity

##### `advanced.head_job_memory`
- **Type**: Integer
- **Optional**: Yes
- **Description**: The memory (in MB) reserved for the Nextflow runner job
- **Format**: Integer in megabytes (MB)
- **Default**: `2048` (2 GB, typical)
- **Constraints**: Multiples of 256MB, from 0.5 GB to 8 GB per CPU
- **Example**: `4096` (4 GB)
- **Character Limit**: 0/6 characters
- **Notes**:
  - Memory for Nextflow orchestration process
  - Large workflows may need more memory
  - Google Cloud enforces memory-to-CPU ratios
  - Example ratios: 0.9 GB to 6.5 GB per vCPU

##### `advanced.service_account_email`
- **Type**: String
- **Optional**: Yes
- **Description**: The service account used when deploying pipeline executions with this compute environment
- **Format**: Email format - `service-account-name@project-id.iam.gserviceaccount.com`
- **Example**: `"pipeline-runner@my-project.iam.gserviceaccount.com"`
- **Character Limit**: 0/200 characters
- **Notes**:
  - Service account must exist in the project
  - Must have necessary permissions:
    - Cloud Batch Job Editor
    - Storage Object Admin (for work bucket)
    - Compute Instance Admin (if creating VMs)
    - Service Account User (if impersonating)
  - If not specified, uses default compute service account

##### `advanced.vpc`
- **Type**: String
- **Optional**: Yes
- **Description**: The name of a VPC network to be used by this compute environment
- **Format**: VPC network name (short name or fully qualified)
- **Examples**:
  - `"my-vpc-network"`
  - `"projects/my-project/global/networks/my-vpc"`
- **Character Limit**: 0/200 characters
- **Notes**:
  - VPC must be accessible to the Google Cloud project
  - Required if using private IP addresses or specific networking
  - Must be in the same project or shared VPC

##### `advanced.subnet`
- **Type**: String
- **Optional**: Yes
- **Description**: The name of a subnet to be used by this compute environment
- **Format**: Subnet name (short name or fully qualified)
- **Examples**:
  - `"my-subnet"`
  - `"projects/my-project/regions/us-central1/subnetworks/my-subnet"`
- **Character Limit**: 0/200 characters
- **Notes**:
  - Must be accessible by the VPC network specified above
  - Must be in the same location as the compute environment
  - Subnet provides IP address range for VMs

##### `advanced.head_job_instance_template`
- **Type**: String
- **Optional**: Yes
- **Description**: The name or fully qualified reference of the instance template to use for the head job
- **Format**:
  - Short name: `"head-template"`
  - Fully qualified: `"projects/my-project/global/instanceTemplates/head-template"`
- **Example**: `"projects/my-project/global/instanceTemplates/nextflow-head"`
- **Character Limit**: 0/400 characters
- **Notes**:
  - Instance template defines machine type, boot disk, network config
  - Overrides individual settings like CPUs and memory
  - Template must exist in the project
  - Useful for standardized configurations

##### `advanced.compute_jobs_instance_template`
- **Type**: String
- **Optional**: Yes
- **Description**: The name or fully qualified reference of the instance template to use for compute tasks
- **Format**:
  - Short name: `"compute-template"`
  - Fully qualified: `"projects/my-project/global/instanceTemplates/compute-template"`
- **Example**: `"projects/my-project/global/instanceTemplates/nextflow-worker"`
- **Character Limit**: 0/400 characters
- **Notes**:
  - Instance template for worker jobs
  - Can define specific machine types, GPUs, disk configurations
  - Useful for consistent worker configurations
  - If not specified, Batch chooses appropriate instance type

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

1. **location**: Must be a valid Google Cloud region
2. **work_directory**: Must start with `gs://` and follow format `gs://bucket/path`
3. **service_account_email**: Must be valid service account email format if specified
4. **head_job_memory**: Must be multiple of 256 MB, within ratio limits per CPU
5. **boot_disk_size**: Must be at least 10 GB
6. **fusion_v2**: Requires `wave.enabled = true`
7. **instance_templates**: Must be valid template references if specified
8. **resource_labels**: Each label must have both `name` and `value` fields

### Lifecycle Considerations

- **Create**: Provisions Google Batch compute environment configuration in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields (some fields may require replacement)
- **Delete**: Removes compute environment from Seqera Platform (terminates any running jobs)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `location`
- `credentials_id`
- `advanced.vpc`
- `advanced.subnet`

### Mutable Fields

These fields can be updated without replacement:
- `work_directory`
- `spot`
- `wave` configuration
- `fusion_v2` configuration
- `resource_labels`
- `staging_options`
- `nextflow_config`
- `environment_variables`
- `advanced.use_private_address`
- `advanced.boot_disk_size`
- `advanced.head_job_cpus`
- `advanced.head_job_memory`
- `advanced.service_account_email`
- `advanced.head_job_instance_template`
- `advanced.compute_jobs_instance_template`

### Sensitive Fields

- The referenced `credentials_id` points to sensitive Google Cloud credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration

```hcl
resource "seqera_compute_google_batch" "minimal" {
  name           = "google-batch-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://my-bucket/work"
}
```

### Cost-Optimized with Spot VMs

```hcl
resource "seqera_compute_google_batch" "spot" {
  name           = "google-batch-spot"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://my-bucket/work"

  spot = true

  nextflow_config = <<-EOF
    process {
      errorStrategy = 'retry'
      maxRetries = 3
      maxErrors = -1
    }
  EOF

  resource_labels = [
    {
      name  = "cost_optimization"
      value = "spot"
    }
  ]
}
```

### Production with Wave and Fusion

```hcl
resource "seqera_compute_google_batch" "production" {
  name           = "google-batch-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://prod-bucket/nextflow-work"

  spot = false

  wave {
    enabled          = true
    strategy         = "conda,container"
    freeze_mode      = true
    build_repository = "us-docker.pkg.dev/my-project/wave/builds"
    cache_repository = "us-docker.pkg.dev/my-project/wave/cache"
  }

  fusion_v2 {
    enabled          = true
    fusion_log_level = "INFO"
  }

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
      export GCP_ENV=production

      # Download reference data
      gsutil -m cp -r gs://reference-bucket/genome /mnt/reference/

      # Validate environment
      echo "Environment ready"
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed with exit status: $NXF_EXIT_STATUS"

      # Archive results
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      gsutil -m cp -r /tmp/results gs://archive-bucket/results-$TIMESTAMP/

      # Send notification
      gcloud pubsub topics publish pipeline-complete \
        --message "Pipeline completed at $TIMESTAMP"
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'google-batch'
      errorStrategy = 'retry'
      maxRetries = 2

      cpus = 4
      memory = '16 GB'

      withLabel: intensive {
        cpus = 16
        memory = '64 GB'
      }
    }

    google {
      project = var.gcp_project_id
      location = 'us-central1'
      batch.spot = false
    }
  EOF

  environment_variables = {
    "NXF_ANSI_LOG"         = "false"
    "NXF_OPTS"             = "-Xms2g -Xmx8g"
    "GOOGLE_CLOUD_PROJECT" = var.gcp_project_id
  }

  advanced {
    boot_disk_size        = 100
    head_job_cpus         = 4
    head_job_memory       = 8192
    service_account_email = "pipeline-runner@my-project.iam.gserviceaccount.com"
  }
}
```

### Private VPC Configuration

```hcl
resource "seqera_compute_google_batch" "private" {
  name           = "google-batch-private"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://private-bucket/work"

  resource_labels = [
    {
      name  = "security"
      value = "private"
    }
  ]

  advanced {
    use_private_address   = true
    vpc                   = "projects/my-project/global/networks/private-vpc"
    subnet                = "projects/my-project/regions/us-central1/subnetworks/private-subnet"
    service_account_email = "private-pipeline@my-project.iam.gserviceaccount.com"
    boot_disk_size        = 50
  }
}
```

### Custom Instance Templates

```hcl
resource "seqera_compute_google_batch" "custom_templates" {
  name           = "google-batch-custom"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://my-bucket/work"

  spot = true

  advanced {
    head_job_instance_template      = "projects/my-project/global/instanceTemplates/nextflow-head-n2-standard-4"
    compute_jobs_instance_template  = "projects/my-project/global/instanceTemplates/nextflow-worker-n2-highmem-8"
    service_account_email           = "pipeline-runner@my-project.iam.gserviceaccount.com"
  }

  resource_labels = [
    {
      name  = "template_type"
      value = "custom"
    }
  ]
}
```

### GPU-Enabled Configuration

```hcl
resource "seqera_compute_google_batch" "gpu" {
  name           = "google-batch-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://gpu-bucket/work"

  resource_labels = [
    {
      name  = "compute_type"
      value = "gpu"
    }
  ]

  nextflow_config = <<-EOF
    process {
      withLabel: gpu {
        accelerator = [request: 1, type: 'nvidia-tesla-t4']
        machineType = 'n1-standard-4'
      }
    }
  EOF

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0"
  }

  advanced {
    boot_disk_size = 100
  }
}
```

### High-Resource Head Job

```hcl
resource "seqera_compute_google_batch" "large_head" {
  name           = "google-batch-large-head"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://my-bucket/work"

  advanced {
    head_job_cpus   = 8
    head_job_memory = 32768  # 32 GB in MB
    boot_disk_size  = 200
  }

  resource_labels = [
    {
      name  = "head_size"
      value = "large"
    }
  ]
}
```

### Multi-Region Pattern

```hcl
# Variables for multi-region deployment
variable "gcp_regions" {
  type = map(object({
    bucket = string
  }))
  default = {
    us-central1 = {
      bucket = "us-central1-bucket"
    }
    europe-west1 = {
      bucket = "europe-west1-bucket"
    }
    asia-southeast1 = {
      bucket = "asia-southeast1-bucket"
    }
  }
}

# Create compute environments in multiple regions
resource "seqera_compute_google_batch" "multi_region" {
  for_each = var.gcp_regions

  name           = "google-batch-${each.key}"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = each.key
  work_directory = "gs://${each.value.bucket}/work"

  spot = true

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

  environment_variables = {
    "GCP_REGION" = each.key
  }
}
```

## Integration with Terraform Google Provider

### Complete Example with Infrastructure

```hcl
# Enable required APIs
resource "google_project_service" "batch" {
  service = "batch.googleapis.com"
}

resource "google_project_service" "compute" {
  service = "compute.googleapis.com"
}

# GCS Bucket for work directory
resource "google_storage_bucket" "work" {
  name     = "my-nextflow-work-bucket"
  location = "US-CENTRAL1"

  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }
}

# Service Account for pipeline execution
resource "google_service_account" "pipeline" {
  account_id   = "pipeline-runner"
  display_name = "Nextflow Pipeline Runner"
}

# IAM permissions for service account
resource "google_project_iam_member" "batch_agent" {
  project = var.gcp_project_id
  role    = "roles/batch.agentReporter"
  member  = "serviceAccount:${google_service_account.pipeline.email}"
}

resource "google_storage_bucket_iam_member" "work_admin" {
  bucket = google_storage_bucket.work.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.pipeline.email}"
}

# VPC Network
resource "google_compute_network" "vpc" {
  name                    = "nextflow-vpc"
  auto_create_subnetworks = false
}

# Subnet
resource "google_compute_subnetwork" "subnet" {
  name          = "nextflow-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = "us-central1"
  network       = google_compute_network.vpc.id

  private_ip_google_access = true
}

# Cloud NAT for private VMs
resource "google_compute_router" "router" {
  name    = "nextflow-router"
  region  = "us-central1"
  network = google_compute_network.vpc.id
}

resource "google_compute_router_nat" "nat" {
  name   = "nextflow-nat"
  router = google_compute_router.router.name
  region = "us-central1"

  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

# Instance Template for compute jobs
resource "google_compute_instance_template" "worker" {
  name_prefix  = "nextflow-worker-"
  machine_type = "n2-standard-4"

  disk {
    source_image = "projects/debian-cloud/global/images/family/debian-11"
    auto_delete  = true
    boot         = true
    disk_size_gb = 50
  }

  network_interface {
    network    = google_compute_network.vpc.id
    subnetwork = google_compute_subnetwork.subnet.id
  }

  service_account {
    email  = google_service_account.pipeline.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    enable-oslogin = "TRUE"
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Seqera Compute Environment
resource "seqera_compute_google_batch" "integrated" {
  name           = "google-batch-integrated"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://${google_storage_bucket.work.name}/work"

  spot = true

  wave {
    enabled = true
  }

  fusion_v2 {
    enabled = true
  }

  advanced {
    use_private_address            = true
    vpc                            = google_compute_network.vpc.name
    subnet                         = google_compute_subnetwork.subnet.name
    service_account_email          = google_service_account.pipeline.email
    compute_jobs_instance_template = google_compute_instance_template.worker.self_link
    boot_disk_size                 = 50
  }

  depends_on = [
    google_project_service.batch,
    google_project_service.compute,
  ]
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
- Resource `location` → API `config.region` or `config.location`
- Resource `work_directory` → API `config.workDir`
- Resource `spot` → API `config.spot`
- Resource `wave` → API `config.wave`
- Resource `fusion_v2` → API `config.fusion2`
- Resource `resource_labels` → API `config.resourceLabels`
- Resource `staging_options.pre_run_script` → API `config.preRunScript`
- Resource `staging_options.post_run_script` → API `config.postRunScript`
- Resource `nextflow_config` → API `config.environment` or `config.nextflowConfig`
- Resource `advanced.use_private_address` → API `config.usePrivateAddress`
- Resource `advanced.boot_disk_size` → API `config.bootDiskSizeGb`
- Resource `advanced.head_job_cpus` → API `config.headJobCpus`
- Resource `advanced.head_job_memory` → API `config.headJobMemoryMb`
- Resource `advanced.service_account_email` → API `config.serviceAccountEmail`
- Resource `advanced.vpc` → API `config.network`
- Resource `advanced.subnet` → API `config.subnetwork`

### Platform Type

- **API Value**: `"google-batch"`
- **Config Type**: `"GoogleBatchComputeConfig"` or `"GCPBatchComputeConfig"`

## Related Resources

- `seqera_google_credential` - Google Cloud credentials used by the compute environment
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Google Cloud Batch Requirements

### Required Permissions

The Google Cloud credentials (service account) must have the following permissions:

#### Batch API Permissions
- `batch.jobs.create`
- `batch.jobs.get`
- `batch.jobs.list`
- `batch.jobs.delete`
- `batch.tasks.list`

#### Compute Engine Permissions
- `compute.instances.create`
- `compute.instances.get`
- `compute.instances.list`
- `compute.instances.delete`
- `compute.machineTypes.get`
- `compute.zones.get`

#### Storage Permissions
- `storage.buckets.get`
- `storage.objects.create`
- `storage.objects.delete`
- `storage.objects.get`
- `storage.objects.list`

#### IAM Permissions
- `iam.serviceAccounts.actAs` (if using service account)

### Required IAM Roles

Recommended IAM roles for the service account:
- **Batch Agent Reporter**: `roles/batch.agentReporter`
- **Batch Job Editor**: `roles/batch.jobsEditor`
- **Storage Object Admin**: `roles/storage.objectAdmin` (for work bucket)
- **Compute Instance Admin**: `roles/compute.instanceAdmin.v1` (if needed)
- **Service Account User**: `roles/iam.serviceAccountUser` (if impersonating)

### Google Cloud Project Configuration

- **APIs Enabled**:
  - Cloud Batch API (`batch.googleapis.com`)
  - Compute Engine API (`compute.googleapis.com`)
  - Cloud Storage API (`storage.googleapis.com`)
- **Quotas**: Sufficient quotas for:
  - Batch jobs
  - Compute Engine CPUs
  - Persistent disk
  - IP addresses (if not using private addresses)
- **Networking**: VPC and subnet configuration if using private IPs

## Best Practices

### Cost Optimization

1. **Use Spot VMs**: Enable `spot = true` for fault-tolerant workloads
2. **Lifecycle Policies**: Set GCS bucket lifecycle rules to delete old work files
3. **Right-size Resources**: Use appropriate head job and worker resources
4. **Regional Bucket**: Place work bucket in same region as compute environment
5. **Preemptible Strategy**: Configure proper retry strategy with Spot VMs

### Security

1. **Private IPs**: Use `use_private_address = true` with Cloud NAT
2. **Service Account**: Use dedicated service account with minimal permissions
3. **VPC Configuration**: Deploy in custom VPC with firewall rules
4. **Secret Management**: Use Secret Manager for sensitive values
5. **IAM Best Practices**: Follow principle of least privilege

### Performance

1. **Fusion v2**: Enable for large file access patterns
2. **Wave Containers**: Use for private container repos and Conda
3. **Regional Colocation**: Keep bucket, VMs in same region
4. **Boot Disk Size**: Increase for large container images
5. **Instance Templates**: Use optimized templates for consistent performance

### Reliability

1. **Error Strategy**: Configure retry logic in Nextflow config
2. **Spot VMs**: Use with proper retry and checkpoint strategies
3. **Resource Labels**: Tag resources for tracking and debugging
4. **Monitoring**: Enable Cloud Monitoring and Logging
5. **Health Checks**: Implement workflow health validation

## References

- [Google Cloud Batch Documentation](https://cloud.google.com/batch/docs)
- [Nextflow Google Cloud Documentation](https://www.nextflow.io/docs/latest/google.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Google Cloud Machine Types](https://cloud.google.com/compute/docs/machine-types)
- [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
- [Spot (Preemptible) VMs](https://cloud.google.com/compute/docs/instances/preemptible)
