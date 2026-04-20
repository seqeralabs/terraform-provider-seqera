# Google Cloud Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_google_cloud` resource, which manages Google Cloud (Compute Engine-based) compute environments in Seqera Platform.

Google Cloud compute environments provide compute capacity using direct Google Compute Engine VMs managed by Seqera Platform, offering a simpler alternative to Google Batch for straightforward workloads.

## Key Differences: Google Cloud vs Google Batch

| Feature | Google Cloud | Google Batch |
|---------|--------------|--------------|
| **Compute Service** | Direct Compute Engine VMs | Google Cloud Batch service |
| **Management** | Seqera manages VMs | Google Batch manages jobs |
| **Configuration** | Simpler setup | More configuration options |
| **Scaling** | Basic VM provisioning | Batch job orchestration |
| **Features** | Core features | Wave, Fusion v2, Spot VMs, instance templates |
| **Control** | Direct VM control | Batch abstraction |
| **Use Case** | Simple workflows, quick setup | Complex batch processing, advanced features |

## Resource Structure

```hcl
resource "seqera_compute_google_cloud" "example" {
  name         = "google-cloud-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"

  work_directory = "gs://my-bucket/work"

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
      executor = 'google-lifesciences'
    }
  EOF

  # Environment variables
  environment_variables = {
    "MY_VAR"     = "value"
    "GCP_REGION" = "us-central1"
  }

  # Advanced options
  advanced {
    use_arm64_architecture = false
    use_gpu_enabled        = false
    instance_type          = "e2-standard-4"
    image_id               = "projects/debian-cloud/global/images/family/debian-11"
    boot_disk_size         = 50
    zone                   = "us-central1-a"
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
- **Example**: `"google-cloud-prod"`

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
  - Create and manage Compute Engine instances
  - Access Google Cloud Storage
  - Manage networking resources

#### `region`
- **Type**: String
- **Required**: Yes
- **Description**: The region where the compute environment executions will happen
- **Validation**: Must be a valid Google Cloud region
- **Examples**: `"us-central1"`, `"europe-west1"`, `"asia-southeast1"`, `"us-east1"`
- **Character Limit**: 0/100 characters
- **Notes**: All resources (VMs, storage) should be in this region for best performance

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: Google Storage bucket to be used as the work directory
- **Format**: `gs://bucket-name/path`
- **Example**: `"gs://my-nextflow-bucket/work"`
- **Character Limit**: 0/400 characters
- **Constraints**:
  - Must start with `gs://`
  - Bucket must exist
  - Credentials must have read/write access
  - **Must be located in the same region chosen previously**

### Optional Fields

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
      value = "genomics-workspace"
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
  - Labels applied to Compute Engine VMs and related resources
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
  gsutil -m cp -r gs://reference-bucket/genome /mnt/reference/

  # Install additional tools
  apt-get update && apt-get install -y jq
  ```
- **Use Cases**:
  - Load environment modules
  - Download reference data
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

  # Cleanup
  rm -rf /tmp/work/*
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
- **Description**: Global Nextflow configuration settings for all pipelines launched with this compute environment
- **Format**: Nextflow configuration DSL
- **Character Limit**: 0/3200 characters
- **Example**:
  ```groovy
  process {
    executor = 'google-lifesciences'

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
      machineType = 'n1-standard-4'
      accelerator = [request: 1, type: 'nvidia-tesla-t4']
    }
  }

  google {
    project = 'my-project-id'
    zone = 'us-central1-a'
    lifeSciences {
      bootDiskSize = '50 GB'
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

##### `advanced.use_arm64_architecture`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable this option to deploy an ARM64 architecture instance
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - ARM64 instances (e.g., `c4a-*` series) offer better price/performance
  - Requires ARM64-compatible container images
  - Default instance type changes to `c4a-standard-4` when enabled
  - Not all software is compatible with ARM64

##### `advanced.use_gpu_enabled`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable this option to deploy a GPU-enabled instance
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - GPU instances for machine learning and compute-intensive workloads
  - Default instance type changes to `a2-highgpu-1g` when enabled
  - Significantly higher cost than CPU-only instances
  - Requires GPU-compatible container images and CUDA libraries

##### `advanced.instance_type`
- **Type**: String
- **Optional**: Yes
- **Description**: Specify the instance type to use
- **Format**: Must be a valid Google machine type
- **Default**:
  - AMD64 (default): `e2-standard-4`
  - ARM64 architecture: `c4a-standard-4`
  - GPU-enabled: `a2-highgpu-1g`
- **Examples**:
  - General Purpose E2: `"e2-standard-2"`, `"e2-standard-4"`, `"e2-standard-8"`, `"e2-standard-16"`
  - General Purpose N2: `"n2-standard-4"`, `"n2-standard-8"`, `"n2-standard-16"`, `"n2-standard-32"`
  - Compute Optimized C2: `"c2-standard-4"`, `"c2-standard-8"`, `"c2-standard-16"`
  - Compute Optimized C3: `"c3-standard-4"`, `"c3-standard-8"`, `"c3-standard-22"`
  - Memory Optimized M2: `"m2-ultramem-208"`, `"m2-ultramem-416"`
  - ARM64 C4A: `"c4a-standard-4"`, `"c4a-standard-8"`, `"c4a-standard-16"`
  - GPU A2: `"a2-highgpu-1g"`, `"a2-highgpu-2g"`, `"a2-highgpu-4g"`, `"a2-ultragpu-1g"`
- **Notes**:
  - Must be available in the specified region and zone
  - Consider CPU, memory, and cost requirements
  - GPU instances only in specific zones

##### `advanced.image_id`
- **Type**: String
- **Optional**: Yes
- **Description**: Custom VM image ID
- **Format**: Image family or specific image URL
- **Examples**:
  - Image family: `"projects/debian-cloud/global/images/family/debian-11"`
  - Specific image: `"projects/debian-cloud/global/images/debian-11-bullseye-v20240213"`
  - Custom image: `"projects/my-project/global/images/my-custom-image"`
- **Notes**:
  - Default image provided by Seqera if not specified
  - Custom images must have required dependencies (Docker, gcloud CLI, etc.)
  - Use for standardized environments or specialized tools

##### `advanced.boot_disk_size`
- **Type**: Integer
- **Optional**: Yes
- **Description**: Enter the boot disk size in GB
- **Default**: `50` (typical default)
- **Range**: Minimum 10 GB
- **Example**: `100`
- **Notes**:
  - Larger disk for container images and temporary files
  - Consider workflow storage requirements
  - Standard persistent disk used by default

##### `advanced.zone`
- **Type**: String
- **Optional**: Yes
- **Description**: The specific zone within the selected region, where the workload will be executed
- **Format**: Zone identifier (region + zone letter)
- **Examples**: `"us-central1-a"`, `"us-central1-b"`, `"europe-west1-b"`, `"asia-southeast1-a"`
- **Notes**:
  - Must be a zone within the specified region
  - Some instance types (especially GPUs) only available in specific zones
  - If not specified, Seqera may choose an appropriate zone
  - Consider availability and quotas

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

1. **region**: Must be a valid Google Cloud region
2. **work_directory**: Must start with `gs://` and be in the same region
3. **instance_type**: Must be a valid Google Compute Engine machine type
4. **zone**: Must be a zone within the specified region
5. **boot_disk_size**: Must be at least 10 GB
6. **Architecture flags**: `use_arm64_architecture` and `use_gpu_enabled` are mutually exclusive in some cases
7. **resource_labels**: Each label must have both `name` and `value` fields

### Lifecycle Considerations

- **Create**: Provisions Google Cloud compute environment configuration in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields (some fields may require replacement)
- **Delete**: Removes compute environment from Seqera Platform (terminates any running VMs)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `region`
- `credentials_id`

### Mutable Fields

These fields can be updated without replacement:
- `work_directory`
- `resource_labels`
- `staging_options` (pre_run_script, post_run_script)
- `nextflow_config`
- `environment_variables`
- `advanced.use_arm64_architecture`
- `advanced.use_gpu_enabled`
- `advanced.instance_type`
- `advanced.image_id`
- `advanced.boot_disk_size`
- `advanced.zone`

### Sensitive Fields

- The referenced `credentials_id` points to sensitive Google Cloud credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration

```hcl
resource "seqera_compute_google_cloud" "minimal" {
  name           = "google-cloud-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://my-bucket/work"
}
```

### Standard Configuration with Resource Labels

```hcl
resource "seqera_compute_google_cloud" "standard" {
  name           = "google-cloud-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://my-work-bucket/nextflow"

  resource_labels = [
    {
      name  = "seqera_workspace"
      value = "genomics"
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
    instance_type  = "n2-standard-8"
    boot_disk_size = 100
    zone           = "us-central1-a"
  }
}
```

### ARM64 Architecture Configuration

```hcl
resource "seqera_compute_google_cloud" "arm64" {
  name           = "google-cloud-arm64"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://my-bucket/work"

  resource_labels = [
    {
      name  = "architecture"
      value = "arm64"
    }
  ]

  advanced {
    use_arm64_architecture = true
    instance_type          = "c4a-standard-8"
    boot_disk_size         = 50
    zone                   = "us-central1-a"
  }

  environment_variables = {
    "ARCH" = "arm64"
  }
}
```

### GPU-Enabled Configuration

```hcl
resource "seqera_compute_google_cloud" "gpu" {
  name           = "google-cloud-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://gpu-bucket/work"

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
    use_gpu_enabled = true
    instance_type   = "a2-highgpu-1g"
    boot_disk_size  = 200
    zone            = "us-central1-a"
  }

  nextflow_config = <<-EOF
    process {
      withLabel: gpu {
        machineType = 'a2-highgpu-1g'
        accelerator = [request: 1, type: 'nvidia-tesla-a100']
      }
    }
  EOF

  environment_variables = {
    "CUDA_VISIBLE_DEVICES"    = "0"
    "NVIDIA_VISIBLE_DEVICES"  = "all"
  }
}
```

### Production Configuration with All Options

```hcl
resource "seqera_compute_google_cloud" "production" {
  name           = "google-cloud-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://prod-bucket/nextflow-work"

  resource_labels = [
    {
      name  = "seqera_workspace"
      value = "production"
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
      export GCP_ENV=production

      # Load required modules
      echo "Installing dependencies..."
      apt-get update && apt-get install -y jq parallel

      # Download reference data from GCS
      echo "Downloading reference data..."
      gsutil -m cp -r gs://reference-bucket/genome /mnt/reference/

      # Validate environment
      echo "Validating environment..."
      gcloud --version
      gsutil --version

      echo "Environment ready"
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed with exit status: $NXF_EXIT_STATUS"

      # Archive results with timestamp
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      ARCHIVE_PATH="results/run-$TIMESTAMP"

      echo "Archiving results to gs://archive-bucket/$ARCHIVE_PATH..."
      gsutil -m cp -r /tmp/results gs://archive-bucket/$ARCHIVE_PATH/

      # Generate summary report
      echo "Generating summary report..."
      cat > /tmp/report.txt <<REPORT
      Pipeline Execution Summary
      ==========================
      Timestamp: $TIMESTAMP
      Exit Status: $NXF_EXIT_STATUS
      Archive Location: gs://archive-bucket/$ARCHIVE_PATH
REPORT

      gsutil cp /tmp/report.txt gs://reports-bucket/report-$TIMESTAMP.txt

      # Send notification via Pub/Sub
      gcloud pubsub topics publish pipeline-complete \
        --message "Pipeline completed at $TIMESTAMP with status $NXF_EXIT_STATUS"

      # Cleanup
      echo "Cleaning up temporary files..."
      rm -rf /tmp/work/*
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'google-lifesciences'
      errorStrategy = 'retry'
      maxRetries = 2

      cpus = 4
      memory = '16 GB'
      disk = '50 GB'

      withLabel: small_task {
        cpus = 2
        memory = '8 GB'
      }

      withLabel: intensive {
        cpus = 16
        memory = '64 GB'
        disk = '200 GB'
      }

      withLabel: high_mem {
        memory = '128 GB'
      }
    }

    google {
      project = var.gcp_project_id
      zone = 'us-central1-a'
      lifeSciences {
        bootDiskSize = '100 GB'
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
    "NXF_ANSI_LOG"         = "false"
    "NXF_OPTS"             = "-Xms2g -Xmx8g"
    "GOOGLE_CLOUD_PROJECT" = var.gcp_project_id
    "GCP_REGION"           = "us-central1"
    "TMPDIR"               = "/tmp"
    "JAVA_HOME"            = "/usr/lib/jvm/java-11"
  }

  advanced {
    instance_type  = "n2-standard-16"
    boot_disk_size = 200
    zone           = "us-central1-a"
    image_id       = "projects/my-project/global/images/custom-nextflow-image"
  }
}
```

### High-Performance Compute Configuration

```hcl
resource "seqera_compute_google_cloud" "high_performance" {
  name           = "google-cloud-hpc"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://hpc-bucket/work"

  resource_labels = [
    {
      name  = "workload_type"
      value = "high_performance"
    }
  ]

  advanced {
    instance_type  = "c3-standard-22"  # Compute optimized
    boot_disk_size = 100
    zone           = "us-central1-a"
  }

  nextflow_config = <<-EOF
    process {
      cpus = 22
      memory = '88 GB'

      withLabel: parallel {
        maxForks = 16
      }
    }
  EOF

  environment_variables = {
    "NXF_OPTS"        = "-Xms4g -Xmx16g"
    "OMP_NUM_THREADS" = "22"
  }
}
```

### Development/Testing Configuration

```hcl
resource "seqera_compute_google_cloud" "dev" {
  name           = "google-cloud-dev"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://dev-bucket/work"

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
    instance_type  = "e2-standard-4"  # Cost-effective
    boot_disk_size = 50
    zone           = "us-central1-b"
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
variable "gcp_regions" {
  type = map(object({
    bucket = string
    zone   = string
  }))
  default = {
    us-central1 = {
      bucket = "us-central1-bucket"
      zone   = "us-central1-a"
    }
    europe-west1 = {
      bucket = "europe-west1-bucket"
      zone   = "europe-west1-b"
    }
    asia-southeast1 = {
      bucket = "asia-southeast1-bucket"
      zone   = "asia-southeast1-a"
    }
  }
}

# Create compute environments in multiple regions
resource "seqera_compute_google_cloud" "multi_region" {
  for_each = var.gcp_regions

  name           = "google-cloud-${each.key}"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = each.key
  work_directory = "gs://${each.value.bucket}/work"

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
    zone = each.value.zone
  }

  environment_variables = {
    "GCP_REGION" = each.key
  }
}
```

## Integration with Terraform Google Provider

### Complete Example with Infrastructure

```hcl
# Enable required APIs
resource "google_project_service" "compute" {
  service = "compute.googleapis.com"
}

resource "google_project_service" "lifesciences" {
  service = "lifesciences.googleapis.com"
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

  labels = {
    purpose     = "nextflow-work"
    environment = "production"
  }
}

# Service Account for pipeline execution
resource "google_service_account" "pipeline" {
  account_id   = "nextflow-pipeline"
  display_name = "Nextflow Pipeline Runner"
}

# IAM permissions for service account
resource "google_project_iam_member" "compute_admin" {
  project = var.gcp_project_id
  role    = "roles/compute.instanceAdmin.v1"
  member  = "serviceAccount:${google_service_account.pipeline.email}"
}

resource "google_storage_bucket_iam_member" "work_admin" {
  bucket = google_storage_bucket.work.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.pipeline.email}"
}

resource "google_project_iam_member" "lifesciences_admin" {
  project = var.gcp_project_id
  role    = "roles/lifesciences.workflowsRunner"
  member  = "serviceAccount:${google_service_account.pipeline.email}"
}

# VPC Network (optional)
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

# Firewall rules
resource "google_compute_firewall" "allow_internal" {
  name    = "nextflow-allow-internal"
  network = google_compute_network.vpc.name

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }

  source_ranges = ["10.0.0.0/24"]
}

# Custom VM Image (optional)
resource "google_compute_image" "custom" {
  name = "nextflow-custom-image"

  source_disk = google_compute_disk.image_source.self_link

  labels = {
    purpose = "nextflow"
  }
}

# Seqera Compute Environment
resource "seqera_compute_google_cloud" "integrated" {
  name           = "google-cloud-integrated"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://${google_storage_bucket.work.name}/work"

  resource_labels = [
    {
      name  = "managed_by"
      value = "terraform"
    },
    {
      name  = "environment"
      value = "production"
    }
  ]

  advanced {
    instance_type  = "n2-standard-8"
    boot_disk_size = 100
    zone           = "us-central1-a"
    image_id       = google_compute_image.custom.self_link
  }

  depends_on = [
    google_project_service.compute,
    google_project_service.lifesciences,
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
- Resource `region` → API `config.region`
- Resource `work_directory` → API `config.workDir`
- Resource `resource_labels` → API `config.resourceLabels`
- Resource `staging_options.pre_run_script` → API `config.preRunScript`
- Resource `staging_options.post_run_script` → API `config.postRunScript`
- Resource `nextflow_config` → API `config.environment` or `config.nextflowConfig`
- Resource `advanced.use_arm64_architecture` → API `config.useArm64Architecture`
- Resource `advanced.use_gpu_enabled` → API `config.useGpuEnabled`
- Resource `advanced.instance_type` → API `config.instanceType` or `config.machineType`
- Resource `advanced.image_id` → API `config.imageId`
- Resource `advanced.boot_disk_size` → API `config.bootDiskSizeGb`
- Resource `advanced.zone` → API `config.zone`

### Platform Type

- **API Value**: `"google-cloud"` or `"gls"` (Google Life Sciences)
- **Config Type**: `"GoogleCloudComputeConfig"` or `"GLSComputeConfig"`

## Related Resources

- `seqera_google_credential` - Google Cloud credentials used by the compute environment
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Comparison: Google Cloud vs Google Batch

| Aspect | Google Cloud | Google Batch |
|--------|--------------|--------------|
| **Setup Complexity** | Simpler | More complex |
| **Configuration** | Minimal options | Extensive options |
| **VM Management** | Direct Compute Engine VMs | Batch job orchestration |
| **Features** | Core features | Spot VMs, Wave, Fusion v2, instance templates |
| **Autoscaling** | Basic | Advanced |
| **Control Level** | Moderate | High |
| **Cost** | Direct VM costs | Batch service features |
| **Use Case** | Simple workflows, quick setup | Complex workflows, advanced features, cost optimization |

## Google Cloud Project Requirements

### Required Permissions

The Google Cloud credentials (service account) must have the following permissions:

#### Compute Engine Permissions
- `compute.instances.create`
- `compute.instances.get`
- `compute.instances.list`
- `compute.instances.delete`
- `compute.machineTypes.get`
- `compute.zones.get`
- `compute.disks.*`

#### Storage Permissions
- `storage.buckets.get`
- `storage.objects.create`
- `storage.objects.delete`
- `storage.objects.get`
- `storage.objects.list`

#### Life Sciences API Permissions (if using)
- `lifesciences.operations.get`
- `lifesciences.operations.list`
- `lifesciences.workflows.run`

### Required IAM Roles

Recommended IAM roles for the service account:
- **Compute Instance Admin**: `roles/compute.instanceAdmin.v1`
- **Storage Object Admin**: `roles/storage.objectAdmin` (for work bucket)
- **Life Sciences Workflows Runner**: `roles/lifesciences.workflowsRunner` (if using)
- **Service Account User**: `roles/iam.serviceAccountUser` (if needed)

### Google Cloud Project Configuration

- **APIs Enabled**:
  - Compute Engine API (`compute.googleapis.com`)
  - Cloud Storage API (`storage.googleapis.com`)
  - Life Sciences API (`lifesciences.googleapis.com`) - optional
- **Quotas**: Sufficient quotas for:
  - Compute Engine instances
  - CPUs
  - Persistent disk
  - IP addresses
  - GPU quotas (if using GPU instances)
- **Networking**: VPC configuration if custom networking is required

## Migration Guide

### Migrating from Google Batch to Google Cloud

If you're currently using Google Batch and want to simplify to Google Cloud:

```hcl
# Before: Google Batch with advanced features
resource "seqera_compute_google_batch" "old" {
  name           = "google-compute"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  work_directory = "gs://my-bucket/work"

  spot = true

  wave {
    enabled = true
  }

  fusion_v2 {
    enabled = true
  }

  advanced {
    boot_disk_size        = 100
    head_job_cpus         = 4
    head_job_memory       = 8192
    service_account_email = "pipeline@my-project.iam.gserviceaccount.com"
  }
}

# After: Google Cloud (simpler)
resource "seqera_compute_google_cloud" "new" {
  name           = "google-compute"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  region         = "us-central1"
  work_directory = "gs://my-bucket/work"

  advanced {
    instance_type  = "n2-standard-4"
    boot_disk_size = 100
    zone           = "us-central1-a"
  }
}
```

**Key Changes:**
- Remove Batch-specific features (spot, wave, fusion_v2)
- Remove head job configuration
- Remove service account email
- Simplified to direct VM provisioning
- Change `location` to `region` and add `zone`

### When to Choose Google Cloud

Choose Google Cloud when:
- You want simpler configuration
- You don't need Spot VMs
- You don't need Wave or Fusion v2 features
- You have straightforward workloads
- You prefer direct VM control
- You're getting started with Nextflow on GCP

Choose Google Batch when:
- You need cost optimization with Spot VMs
- You want Wave containers or Fusion v2
- You need advanced batch job orchestration
- You have complex workflows
- You want instance templates for standardization
- You need service account customization

## Best Practices

### Cost Optimization

1. **Right-size Instances**: Use appropriate instance types for workload
2. **Lifecycle Policies**: Set GCS bucket lifecycle rules to delete old work files
3. **Regional Resources**: Keep bucket and VMs in same region
4. **E2 Instances**: Use cost-effective E2 series for general workloads
5. **Sustained Use Discounts**: Benefit from automatic discounts for long-running VMs

### Security

1. **Service Account**: Use dedicated service account with minimal permissions
2. **Private Access**: Configure Private Google Access for VPC
3. **Firewall Rules**: Restrict network access appropriately
4. **Secret Management**: Use Secret Manager for sensitive values
5. **IAM Best Practices**: Follow principle of least privilege

### Performance

1. **Regional Colocation**: Keep bucket and VMs in same region
2. **Boot Disk Size**: Increase for large container images
3. **Instance Selection**: Choose appropriate CPU/memory for workload
4. **Network**: Use high-performance instance types if network-intensive

### Reliability

1. **Error Strategy**: Configure retry logic in Nextflow config
2. **Resource Labels**: Tag resources for tracking and debugging
3. **Monitoring**: Enable Cloud Monitoring and Logging
4. **Zonal Availability**: Consider zone availability and quotas
5. **Health Checks**: Implement workflow health validation

## References

- [Google Compute Engine Documentation](https://cloud.google.com/compute/docs)
- [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
- [Nextflow Google Cloud Documentation](https://www.nextflow.io/docs/latest/google.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Google Cloud Machine Types](https://cloud.google.com/compute/docs/machine-types)
- [Google Cloud Regions and Zones](https://cloud.google.com/compute/docs/regions-zones)
