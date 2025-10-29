# Google GKE Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_google_gke` resource, which manages Google GKE (Google Kubernetes Engine) compute environments in Seqera Platform.

Google GKE compute environments enable running Nextflow workflows on Google-managed Kubernetes clusters, combining the benefits of Kubernetes container orchestration with Google Cloud-native integrations.

## Key Characteristics

- **Google-Managed Kubernetes**: Fully managed control plane by Google Cloud
- **GCP Integration**: Native GCS, IAM, VPC, and Cloud Monitoring integration
- **Fusion Storage**: Optional high-performance GCS access via Fusion v2
- **Workload Identity**: Google Cloud IAM integration for pod-level permissions
- **Autopilot Support**: Serverless GKE with automated cluster management
- **Scalability**: Cluster Autoscaler and Vertical Pod Autoscaling support

## Resource Structure

```hcl
resource "seqera_compute_google_gke" "example" {
  name         = "gke-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-west1"

  # GKE cluster
  cluster_name = "my-gke-cluster"
  namespace    = "nextflow"

  # Storage mode
  storage_mode = "fusion"  # or "legacy"

  # Work directory (GCS)
  work_directory = "gs://my-bucket/work"

  # Service accounts
  head_service_account    = "tower-launcher-sa"
  compute_service_account = "default"

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
      executor = 'k8s'
    }
  EOF

  # Environment variables
  environment_variables = {
    "GOOGLE_CLOUD_PROJECT" = "my-project"
  }

  # Advanced options
  advanced {
    pod_cleanup_policy   = "on_success"
    custom_head_pod_specs = <<-YAML
      spec:
        nodeSelector:
          workload: nextflow
        tolerations:
        - key: dedicated
          operator: Equal
          value: nextflow
          effect: NoSchedule
    YAML
    head_job_cpus   = 2
    head_job_memory = 4096
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
- **Example**: `"gke-prod"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: Google Cloud credentials ID to use for accessing GKE and GCP services
- **Reference**: Must reference a valid `seqera_google_credential` resource
- **Example**: `seqera_google_credential.main.credentials_id`
- **Notes**: Credentials must have permissions for:
  - GKE cluster access
  - GCS bucket access
  - IAM (for Workload Identity)
  - Compute Engine (for node management)

#### `location`
- **Type**: String
- **Required**: Yes
- **Description**: Specify the Google region or zone where the cluster located
- **Examples**: `"us-west1"`, `"us-central1"`, `"europe-west1"`, `"asia-southeast1-a"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Can be a region (e.g., `us-west1`) or zone (e.g., `us-west1-a`)
  - GKE cluster must be in this location
  - Regional clusters are highly available

#### `cluster_name`
- **Type**: String
- **Required**: Yes
- **Description**: The GKE cluster name
- **Example**: `"my-gke-cluster"`, `"production-gke"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Cluster must exist in the specified location
  - Credentials must have access to the cluster
  - Can be obtained from GKE console or gcloud CLI

#### `namespace`
- **Type**: String
- **Required**: Yes
- **Description**: The Kubernetes namespace to use for the pipeline execution
- **Example**: `"nextflow"`, `"tower-nf"`, `"default"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Namespace must exist in the GKE cluster
  - Service accounts must have permissions in this namespace

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: The Google storage bucket path to be used as pipeline work directory
- **Format**: `gs://bucket-name/path`
- **Example**: `"gs://my-nextflow-bucket/work"`
- **Character Limit**: 0/200 characters
- **Constraints**:
  - Must start with `gs://`
  - **The Google Storage bucket must be located in the same location entered above**
  - Credentials/service accounts must have read/write access

### Storage Mode

#### `storage_mode`
- **Type**: String
- **Required**: Yes (via radio button selection)
- **Description**: Storage access method for Nextflow work directory
- **Allowed Values**:
  - `"fusion"` - Fusion storage (recommended)
  - `"legacy"` - Legacy storage
- **Default**: `"fusion"` (shown as pre-selected in UI)

##### Fusion Storage Mode
- **Description**: Allow access to your Google Cloud-hosted data via the Fusion v2 virtual distributed file system, speeding up most operations
- **Benefits**:
  - High-performance GCS access
  - Eliminates the need to configure a shared file system in your Kubernetes cluster
  - Lazy loading of files
  - Reduced data transfer costs
  - Improved pipeline performance
- **Notes**: This enables Fusion v2 for optimized GCS access

##### Legacy Storage Mode
- **Description**: For this option, the Nextflow work directory for your data pipeline must be located on a POSIX-compatible file system
- **Requirements**: You must configure a shared file system in your Kubernetes cluster
- **Notes**:
  - Requires NFS, Filestore, or other POSIX-compatible shared storage
  - Mounted via PersistentVolumeClaims
  - Traditional file system semantics

### Optional Fields

#### `head_service_account`
- **Type**: String
- **Optional**: Yes
- **Description**: The Kubernetes service account to connect to the cluster and launch the workflow execution
- **Example**: `"tower-launcher-sa"`, `"nextflow-head"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Service account for the Nextflow launcher/head pod
  - Must exist in the namespace
  - Needs RBAC permissions to create/manage pods
  - Can use Workload Identity for GCP permissions

#### `compute_service_account`
- **Type**: String
- **Optional**: Yes
- **Description**: The service account to use for Nextflow-submitted pipeline jobs
- **Default**: `"default"`
- **Example**: `"tower-job-sa"`, `"nextflow-worker"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Service account for task pods
  - Must exist in the namespace
  - Can use Workload Identity for GCS access without credentials
  - Default is "default" service account

### Resource Labels

#### `resource_labels`
- **Type**: List of Objects
- **Optional**: Yes
- **Description**: Associate name/value pairs with the resources created by this compute environment
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
    }
  ]
  ```
- **Notes**:
  - Only one resource label with the same name can be used (API constraint)
  - Default resource labels are pre-filled
  - Labels applied to Kubernetes pods and GCP resources

### Staging Options Block

#### `staging_options`
Configuration for workflow staging and lifecycle scripts.

##### `staging_options.pre_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script that executes before pipeline launch in the same environment where Nextflow runs
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Setting up GKE environment..."
  export GOOGLE_CLOUD_PROJECT=my-project

  # Verify cluster access
  kubectl get nodes

  # Check GCS access
  gsutil ls gs://my-bucket/

  # Download reference data
  gsutil -m cp -r gs://reference-bucket/genome /mnt/reference/
  ```
- **Use Cases**:
  - Validate GKE cluster connectivity
  - Check GCS access and permissions
  - Download reference data
  - Set up environment

##### `staging_options.post_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script that executes immediately after pipeline completion in the same environment where Nextflow runs
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
    --message "Pipeline completed at $TIMESTAMP"

  # Cleanup
  rm -rf /tmp/work/*
  ```
- **Use Cases**:
  - Archive results to GCS
  - Send notifications via Pub/Sub
  - Cleanup temporary files
  - Generate reports

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
    executor = 'k8s'
    container = 'nextflow/nextflow:latest'

    errorStrategy = 'retry'
    maxRetries = 3

    cpus = 2
    memory = '4 GB'

    withLabel: big_mem {
      memory = '32 GB'
    }
  }

  k8s {
    namespace = 'nextflow'
    serviceAccount = 'tower-job-sa'

    pod {
      nodeSelector = 'workload=nextflow'
    }
  }

  google {
    project = 'my-project-id'
    zone = 'us-west1-a'
  }
  ```
- **Use Cases**:
  - Configure Kubernetes executor
  - Set GCP-specific options
  - Define resource requirements
  - Configure error handling

### Environment Variables

#### `environment_variables`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables set in all workflow pods
- **Example**:
  ```hcl
  environment_variables = {
    "GOOGLE_CLOUD_PROJECT" = "my-project-id"
    "NXF_ANSI_LOG"         = "false"
    "NXF_OPTS"             = "-Xms1g -Xmx4g"
    "FUSION_ENABLED"       = "true"
  }
  ```
- **Notes**:
  - Available to all processes
  - GCP credentials automatically available via Workload Identity if configured

### Advanced Options Block

#### `advanced`
Advanced configuration options for GKE integration.

##### `advanced.pod_cleanup_policy`
- **Type**: String
- **Optional**: Yes
- **Description**: Delete the pod when the job has terminated
- **Allowed Values**:
  - `"on_success"` - Delete pods only when job completes successfully
  - `"always"` - Always delete pods after job completion
  - `"never"` - Never automatically delete pods
- **Default**: `"on_success"`
- **Example**: `"always"`
- **Notes**:
  - Helps manage Kubernetes resource usage
  - `"never"` useful for debugging failed jobs

##### `advanced.custom_head_pod_specs`
- **Type**: String (YAML)
- **Optional**: Yes
- **Description**: Provide a custom configuration for the pod running the Nextflow workflow
- **Format**: Valid PodSpec YAML structure starting with `"spec:"`
- **Example**:
  ```yaml
  spec:
    nodeSelector:
      workload: nextflow
      cloud.google.com/gke-nodepool: nextflow-pool
    tolerations:
    - key: dedicated
      operator: Equal
      value: nextflow
      effect: NoSchedule
    - key: cloud.google.com/gke-spot
      operator: Equal
      value: "true"
      effect: NoSchedule
    affinity:
      nodeAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          preference:
            matchExpressions:
            - key: cloud.google.com/gke-spot
              operator: In
              values:
              - "true"
    securityContext:
      runAsUser: 1000
      fsGroup: 1000
  ```
- **Notes**:
  - Applies to the Nextflow head/launcher pod
  - Can specify nodeSelector, affinity, tolerations, securityContext
  - Must be valid Kubernetes PodSpec YAML
  - Allows for Autopilot, Spot VMs, GPU scheduling

##### `advanced.head_job_cpus`
- **Type**: Integer
- **Optional**: Yes
- **Description**: The number of CPUs to be allocated for the Nextflow runner job
- **Default**: `1` or `2` (typical)
- **Example**: `4`
- **Range**: 1 to node capacity
- **Notes**:
  - CPUs for the head/launcher pod
  - More CPUs for large DAG workflows

##### `advanced.head_job_memory`
- **Type**: Integer
- **Optional**: Yes
- **Description**: The amount of memory in megabytes (MB) reserved for the Nextflow runner job
- **Default**: `2048` (2 GB, typical)
- **Example**: `4096` (4 GB)
- **Format**: Integer in megabytes
- **Notes**:
  - Memory for the head/launcher pod
  - Large workflows may need more memory

## Read-Only Computed Fields

### `compute_env_id`
- **Type**: String
- **Computed**: Yes
- **Description**: Unique identifier for the compute environment
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

1. **location**: Must be a valid Google Cloud region or zone
2. **cluster_name**: Required, cannot be empty
3. **namespace**: Must be valid Kubernetes namespace
4. **work_directory**: Must start with `gs://` and be in the same location
5. **storage_mode**: Must be either "fusion" or "legacy"
6. **pod_cleanup_policy**: Must be one of: on_success, always, never
7. **custom_head_pod_specs**: Must be valid PodSpec YAML starting with "spec:"

### Lifecycle Considerations

- **Create**: Configures GKE compute environment in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields
- **Delete**: Removes compute environment (doesn't delete GKE cluster)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `credentials_id`
- `location`
- `cluster_name`
- `namespace`

### Mutable Fields

These fields can be updated without replacement:
- `storage_mode`
- `work_directory`
- `head_service_account`
- `compute_service_account`
- `resource_labels`
- `staging_options`
- `nextflow_config`
- `environment_variables`
- `advanced` options

### Sensitive Fields

- The referenced `credentials_id` points to sensitive Google Cloud credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration with Fusion

```hcl
resource "seqera_compute_google_gke" "minimal" {
  name           = "gke-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-west1"
  cluster_name   = "my-gke-cluster"
  namespace      = "nextflow"
  storage_mode   = "fusion"
  work_directory = "gs://my-bucket/work"
}
```

### Standard Configuration with Service Accounts

```hcl
resource "seqera_compute_google_gke" "standard" {
  name           = "gke-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  cluster_name   = "production-gke"
  namespace      = "nextflow"
  storage_mode   = "fusion"
  work_directory = "gs://prod-bucket/work"

  head_service_account    = "tower-launcher-sa"
  compute_service_account = "tower-job-sa"

  resource_labels = [
    {
      name  = "environment"
      value = "production"
    },
    {
      name  = "cost_center"
      value = "genomics"
    }
  ]
}
```

### Legacy Storage with Filestore

```hcl
resource "seqera_compute_google_gke" "legacy" {
  name           = "gke-legacy-filestore"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-west1"
  cluster_name   = "my-gke-cluster"
  namespace      = "nextflow"
  storage_mode   = "legacy"
  work_directory = "/mnt/filestore/nextflow-work"

  # Note: Requires Filestore CSI driver and PVC configured in cluster
  resource_labels = [
    {
      name  = "storage_type"
      value = "filestore"
    }
  ]
}
```

### Production with Custom Head Pod Specs

```hcl
resource "seqera_compute_google_gke" "production" {
  name           = "gke-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-west1"
  cluster_name   = "production-gke"
  namespace      = "production-pipelines"
  storage_mode   = "fusion"
  work_directory = "gs://prod-bucket/work"

  head_service_account    = "tower-launcher-prod"
  compute_service_account = "tower-job-prod"

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

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Initializing production GKE environment..."

      # Verify cluster access
      kubectl get nodes --no-headers | wc -l

      # Check GCS access
      gsutil ls gs://prod-bucket/ || exit 1

      # Download reference data
      gsutil -m cp -r gs://reference-bucket/genome /mnt/reference/

      echo "Environment ready"
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed: $NXF_EXIT_STATUS"

      # Archive results
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      gsutil -m cp -r /tmp/results gs://archive-bucket/results-$TIMESTAMP/

      # Send Pub/Sub notification
      gcloud pubsub topics publish pipeline-complete \
        --message "Pipeline completed at $TIMESTAMP"

      # Cleanup
      rm -rf /tmp/work/*
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'k8s'
      errorStrategy = 'retry'
      maxRetries = 2

      cpus = 4
      memory = '16 GB'

      withLabel: intensive {
        cpus = 16
        memory = '64 GB'
      }
    }

    k8s {
      namespace = 'production-pipelines'
      serviceAccount = 'tower-job-prod'

      pod {
        nodeSelector = 'workload=nextflow'
      }
    }

    google {
      project = var.gcp_project_id
      zone = 'us-west1-a'
    }
  EOF

  environment_variables = {
    "GOOGLE_CLOUD_PROJECT" = var.gcp_project_id
    "NXF_ANSI_LOG"         = "false"
    "NXF_OPTS"             = "-Xms2g -Xmx8g"
    "FUSION_ENABLED"       = "true"
  }

  advanced {
    pod_cleanup_policy = "on_success"

    custom_head_pod_specs = <<-YAML
      spec:
        nodeSelector:
          workload: nextflow
          cloud.google.com/gke-nodepool: nextflow-pool
        tolerations:
        - key: dedicated
          operator: Equal
          value: nextflow
          effect: NoSchedule
        securityContext:
          runAsUser: 1000
          fsGroup: 1000
    YAML

    head_job_cpus   = 4
    head_job_memory = 8192
  }
}
```

### Spot VMs Configuration

```hcl
resource "seqera_compute_google_gke" "spot" {
  name           = "gke-spot"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  cluster_name   = "my-gke-cluster"
  namespace      = "nextflow"
  storage_mode   = "fusion"
  work_directory = "gs://my-bucket/work"

  resource_labels = [
    {
      name  = "instance_type"
      value = "spot"
    }
  ]

  nextflow_config = <<-EOF
    process {
      errorStrategy = 'retry'
      maxRetries = 3
      maxErrors = -1
    }

    k8s {
      pod {
        nodeSelector = 'cloud.google.com/gke-spot=true'

        tolerations = [[
          key: 'cloud.google.com/gke-spot',
          operator: 'Equal',
          value: 'true',
          effect: 'NoSchedule'
        ]]
      }
    }
  EOF

  advanced {
    custom_head_pod_specs = <<-YAML
      spec:
        nodeSelector:
          cloud.google.com/gke-spot: "true"
        tolerations:
        - key: cloud.google.com/gke-spot
          operator: Equal
          value: "true"
          effect: NoSchedule
    YAML
  }
}
```

### Autopilot Configuration

```hcl
resource "seqera_compute_google_gke" "autopilot" {
  name           = "gke-autopilot"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-west1"
  cluster_name   = "autopilot-gke-cluster"
  namespace      = "nextflow"
  storage_mode   = "fusion"
  work_directory = "gs://autopilot-bucket/work"

  resource_labels = [
    {
      name  = "cluster_mode"
      value = "autopilot"
    }
  ]

  environment_variables = {
    "GOOGLE_CLOUD_PROJECT" = var.gcp_project_id
  }

  advanced {
    # Autopilot automatically manages resources
    # No custom pod specs needed
    head_job_cpus   = 2
    head_job_memory = 4096
  }
}
```

### GPU-Enabled Configuration

```hcl
resource "seqera_compute_google_gke" "gpu" {
  name           = "gke-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-central1"
  cluster_name   = "gpu-gke-cluster"
  namespace      = "gpu-pipelines"
  storage_mode   = "fusion"
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
        accelerator = 1
        containerOptions = '--gpus all'
      }
    }

    k8s {
      pod {
        nodeSelector = 'cloud.google.com/gke-accelerator=nvidia-tesla-t4'
      }
    }
  EOF

  advanced {
    custom_head_pod_specs = <<-YAML
      spec:
        nodeSelector:
          cloud.google.com/gke-nodepool: standard-pool
    YAML
  }

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0"
  }
}
```

## Integration with Terraform Google Provider

### Complete Example with GKE Infrastructure

```hcl
# VPC for GKE
resource "google_compute_network" "vpc" {
  name                    = "nextflow-gke-vpc"
  auto_create_subnetworks = false
}

# Subnet
resource "google_compute_subnetwork" "subnet" {
  name          = "nextflow-gke-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = "us-west1"
  network       = google_compute_network.vpc.id

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.1.0.0/16"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.2.0.0/16"
  }

  private_ip_google_access = true
}

# GKE Cluster
resource "google_container_cluster" "primary" {
  name     = "nextflow-gke"
  location = "us-west1"

  # Remove default node pool
  remove_default_node_pool = true
  initial_node_count       = 1

  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  workload_identity_config {
    workload_pool = "${var.gcp_project_id}.svc.id.goog"
  }

  release_channel {
    channel = "REGULAR"
  }
}

# Node Pool
resource "google_container_node_pool" "nextflow" {
  name       = "nextflow-pool"
  location   = "us-west1"
  cluster    = google_container_cluster.primary.name
  node_count = 1

  autoscaling {
    min_node_count = 1
    max_node_count = 10
  }

  node_config {
    machine_type = "n2-standard-4"
    disk_size_gb = 100

    labels = {
      workload = "nextflow"
    }

    taint {
      key    = "dedicated"
      value  = "nextflow"
      effect = "NO_SCHEDULE"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }
}

# GCS Bucket for work directory
resource "google_storage_bucket" "work" {
  name     = "nextflow-gke-work-bucket"
  location = "US-WEST1"

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
    purpose = "nextflow-work"
  }
}

# Google Service Account for Workload Identity - Head
resource "google_service_account" "head" {
  account_id   = "nextflow-head"
  display_name = "Nextflow Head Service Account"
}

# Google Service Account for Workload Identity - Worker
resource "google_service_account" "worker" {
  account_id   = "nextflow-worker"
  display_name = "Nextflow Worker Service Account"
}

# IAM binding for GCS access - Head
resource "google_storage_bucket_iam_member" "head" {
  bucket = google_storage_bucket.work.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.head.email}"
}

# IAM binding for GCS access - Worker
resource "google_storage_bucket_iam_member" "worker" {
  bucket = google_storage_bucket.work.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.worker.email}"
}

# Workload Identity binding - Head
resource "google_service_account_iam_member" "head_workload_identity" {
  service_account_id = google_service_account.head.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.gcp_project_id}.svc.id.goog[nextflow/tower-launcher-sa]"
}

# Workload Identity binding - Worker
resource "google_service_account_iam_member" "worker_workload_identity" {
  service_account_id = google_service_account.worker.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.gcp_project_id}.svc.id.goog[nextflow/tower-job-sa]"
}

# Kubernetes Provider
provider "kubernetes" {
  host  = "https://${google_container_cluster.primary.endpoint}"
  token = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(
    google_container_cluster.primary.master_auth[0].cluster_ca_certificate
  )
}

data "google_client_config" "default" {}

# Kubernetes Namespace
resource "kubernetes_namespace" "nextflow" {
  metadata {
    name = "nextflow"

    labels = {
      name = "nextflow"
    }
  }
}

# Service Account for Head Pod
resource "kubernetes_service_account" "head" {
  metadata {
    name      = "tower-launcher-sa"
    namespace = kubernetes_namespace.nextflow.metadata[0].name

    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.head.email
    }
  }
}

# Service Account for Worker Pods
resource "kubernetes_service_account" "worker" {
  metadata {
    name      = "tower-job-sa"
    namespace = kubernetes_namespace.nextflow.metadata[0].name

    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.worker.email
    }
  }
}

# Role for Head Pod
resource "kubernetes_role" "head" {
  metadata {
    name      = "nextflow-head"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  rule {
    api_groups = [""]
    resources  = ["pods", "pods/log", "pods/status"]
    verbs      = ["get", "list", "watch", "create", "delete"]
  }

  rule {
    api_groups = [""]
    resources  = ["configmaps"]
    verbs      = ["get", "list", "create", "delete"]
  }
}

# RoleBinding for Head Pod
resource "kubernetes_role_binding" "head" {
  metadata {
    name      = "nextflow-head"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.head.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.head.metadata[0].name
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }
}

# Role for Worker Pods
resource "kubernetes_role" "worker" {
  metadata {
    name      = "nextflow-worker"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  rule {
    api_groups = [""]
    resources  = ["pods", "pods/log", "pods/status"]
    verbs      = ["get", "list", "watch"]
  }
}

# RoleBinding for Worker Pods
resource "kubernetes_role_binding" "worker" {
  metadata {
    name      = "nextflow-worker"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.worker.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.worker.metadata[0].name
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }
}

# Seqera Compute Environment
resource "seqera_compute_google_gke" "integrated" {
  name           = "gke-integrated"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id
  location       = "us-west1"
  cluster_name   = google_container_cluster.primary.name
  namespace      = kubernetes_namespace.nextflow.metadata[0].name
  storage_mode   = "fusion"
  work_directory = "gs://${google_storage_bucket.work.name}/work"

  head_service_account    = kubernetes_service_account.head.metadata[0].name
  compute_service_account = kubernetes_service_account.worker.metadata[0].name

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
    pod_cleanup_policy = "on_success"

    custom_head_pod_specs = <<-YAML
      spec:
        nodeSelector:
          workload: nextflow
        tolerations:
        - key: dedicated
          operator: Equal
          value: nextflow
          effect: NoSchedule
    YAML

    head_job_cpus   = 2
    head_job_memory = 4096
  }

  depends_on = [
    kubernetes_role_binding.head,
    kubernetes_role_binding.worker,
    google_service_account_iam_member.head_workload_identity,
    google_service_account_iam_member.worker_workload_identity,
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

- Resource `name` → API `config.name`
- Resource `location` → API `config.region` or `config.location`
- Resource `cluster_name` → API `config.clusterName`
- Resource `namespace` → API `config.namespace`
- Resource `storage_mode` → API `config.storageMode` or `config.fusionEnabled`
- Resource `work_directory` → API `config.workDir`
- Resource `head_service_account` → API `config.headServiceAccount`
- Resource `compute_service_account` → API `config.serviceAccount`
- Resource `resource_labels` → API `config.resourceLabels`
- Resource `advanced.pod_cleanup_policy` → API `config.podCleanupPolicy`
- Resource `advanced.custom_head_pod_specs` → API `config.customHeadPodSpecs`
- Resource `advanced.head_job_cpus` → API `config.headJobCpus`
- Resource `advanced.head_job_memory` → API `config.headJobMemoryMb`

### Platform Type

- **API Value**: `"gke"` or `"google-gke"`
- **Config Type**: `"GKEComputeConfig"` or `"GoogleGKEComputeConfig"`

## Related Resources

- `seqera_google_credential` - Google Cloud credentials for GKE and GCS access
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## GKE Cluster Requirements

### Required Google Cloud Permissions

The Google Cloud credentials must have:
- `container.clusters.get`
- `container.clusters.list`
- `storage.objects.get`, `storage.objects.create`, `storage.objects.list` (for GCS bucket)
- `iam.serviceAccounts.getIamPolicy` (if using Workload Identity)

### Required Kubernetes RBAC

Service accounts need appropriate RBAC permissions (see examples above).

### Cluster Configuration

- **Kubernetes Version**: 1.20+
- **Workload Identity**: Enabled for GCP IAM integration
- **VPC**: Private cluster with secondary IP ranges for pods and services
- **Storage**: GCS for Fusion mode, Filestore for Legacy mode
- **Node Pools**: Appropriate machine types and autoscaling

## Storage Options Comparison

| Feature | Fusion Storage | Legacy Storage |
|---------|----------------|----------------|
| **Backend** | GCS via Fusion v2 | POSIX filesystem (Filestore) |
| **Performance** | High (lazy loading) | Depends on filesystem |
| **Setup** | Simpler (just GCS) | Requires Filestore CSI driver + PVC |
| **Cost** | GCS storage costs | Filestore costs (higher) |
| **Scalability** | GCS unlimited | Limited by filesystem |
| **Best For** | Most workflows | POSIX-requiring workflows |

## Best Practices

### Workload Identity

1. **Use Workload Identity**: Avoid GCP credentials in pods
2. **Separate Service Accounts**: Different accounts for head and worker pods
3. **Least Privilege**: Only grant necessary GCS permissions
4. **Multiple Buckets**: Different service accounts for different buckets

### Node Selection

1. **Node Selectors**: Direct workloads to appropriate node pools
2. **Spot VMs**: Use with tolerations for 60-91% cost savings
3. **Autopilot**: Serverless option with automated management
4. **GPU Nodes**: Dedicated node pools for GPU workloads

### Storage

1. **Use Fusion**: Recommended for most workflows
2. **Regional GCS**: Bucket in same location as GKE
3. **Lifecycle Policies**: Delete old work files automatically
4. **Versioning**: Enable for important results buckets

### Security

1. **Private Cluster**: Use private endpoint for GKE API
2. **VPC**: Deploy in private subnets
3. **Firewall Rules**: Restrict network access
4. **Pod Security**: Use SecurityContext and Pod Security Standards
5. **Secrets**: Use Secret Manager or Kubernetes Secrets

### Performance

1. **Fusion v2**: Faster than legacy Filestore
2. **Machine Types**: Choose appropriate compute-optimized types
3. **Cluster Autoscaler**: Auto-scale nodes based on workload
4. **Network**: Ensure high bandwidth between pods and GCS

### Cost Optimization

1. **Spot VMs**: 60-91% cheaper with proper tolerations
2. **Autopilot**: Pay only for pod resources
3. **GCS Lifecycle**: Delete old work files
4. **Right-Sizing**: Use appropriate machine types
5. **Preemptible VMs**: Cost-effective for fault-tolerant workloads

### Monitoring

1. **Cloud Monitoring**: GKE cluster and GCS metrics
2. **GKE Dashboard**: Pod-level metrics
3. **Prometheus**: Kubernetes metrics
4. **Cloud Trace**: Distributed tracing
5. **Cloud Logging**: Centralized logging

## Troubleshooting

### Common Issues

1. **Cannot connect to cluster**:
   - Verify GCP credentials have container.clusters.get permission
   - Check cluster endpoint accessibility
   - Verify kubeconfig/credentials

2. **Pods cannot access GCS**:
   - Verify Workload Identity annotation on service account
   - Check Google Service Account IAM policy
   - Verify GCS bucket permissions

3. **Pods fail to schedule**:
   - Check node selectors and tolerations
   - Verify node capacity
   - Check pod resource requests

4. **Fusion storage not working**:
   - Verify storage_mode = "fusion"
   - Check GCS bucket permissions
   - Verify Fusion v2 is installed in cluster

5. **Permission denied errors**:
   - Check Kubernetes service account RBAC
   - Verify SecurityContext UID/GID
   - Check GCS bucket IAM

## References

- [Google Kubernetes Engine Documentation](https://cloud.google.com/kubernetes-engine/docs)
- [GKE Best Practices](https://cloud.google.com/kubernetes-engine/docs/best-practices)
- [Workload Identity Documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
- [Nextflow Kubernetes Executor](https://www.nextflow.io/docs/latest/kubernetes.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [GKE Autopilot](https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-overview)
- [Spot VMs on GKE](https://cloud.google.com/kubernetes-engine/docs/how-to/spot-vms)
- [Filestore CSI Driver](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/filestore-csi-driver)
