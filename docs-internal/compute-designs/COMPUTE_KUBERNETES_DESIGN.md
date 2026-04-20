# Kubernetes Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_kubernetes` resource, which manages Kubernetes compute environments in Seqera Platform.

Kubernetes compute environments enable running Nextflow workflows on existing Kubernetes clusters, providing container orchestration, resource management, and scalability.

## Key Characteristics

- **Platform-Agnostic**: Works with any Kubernetes cluster (GKE, EKS, AKS, on-premises, etc.)
- **Container-Native**: Leverages Kubernetes pod orchestration
- **Resource Management**: Fine-grained CPU, memory, and storage control
- **Scalability**: Automatic scaling via Kubernetes autoscaling
- **Multi-Cloud**: Can run on any cloud provider or on-premises

## Resource Structure

```hcl
resource "seqera_compute_kubernetes" "example" {
  name         = "kubernetes-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_kubernetes_credential.main.credentials_id

  # Namespace and service account
  namespace          = "nextflow"
  service_account    = "nextflow-runner"

  # Work directory (can be various storage types)
  work_directory = "s3://my-bucket/work"
  # or work_directory = "gs://my-bucket/work"
  # or work_directory = "/mnt/shared/work"

  # Compute resources
  compute_resources {
    cpus   = 4
    memory = "16 GB"
  }

  # Storage configuration
  storage_claims = [
    {
      name        = "shared-data"
      mount_path  = "/mnt/data"
      claim_name  = "shared-pvc"
    }
  ]

  # Pod configuration
  pod_options {
    node_selector = {
      "workload" = "nextflow"
      "tier"     = "compute"
    }

    tolerations = [
      {
        key      = "dedicated"
        operator = "Equal"
        value    = "nextflow"
        effect   = "NoSchedule"
      }
    ]

    labels = {
      "app"     = "nextflow"
      "managed" = "seqera"
    }

    annotations = {
      "prometheus.io/scrape" = "true"
    }
  }

  # Staging options
  staging_options {
    pre_run_script  = "#!/bin/bash\necho 'Starting workflow'"
    post_run_script = "#!/bin/bash\necho 'Workflow complete'"
  }

  # Nextflow configuration
  nextflow_config = <<-EOF
    process {
      executor = 'k8s'
      container = 'nextflow/nextflow:latest'
    }
  EOF

  # Environment variables
  environment_variables = {
    "MY_VAR" = "value"
  }

  # Advanced options
  advanced {
    head_service_account = "nextflow-head"
    image_pull_policy    = "IfNotPresent"
    image_pull_secrets   = ["docker-registry-secret"]
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
- **Example**: `"kubernetes-prod"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: Kubernetes credentials ID to use for accessing the cluster
- **Reference**: Must reference a valid `seqera_kubernetes_credential` resource
- **Example**: `seqera_kubernetes_credential.main.credentials_id`
- **Notes**: Credentials contain kubeconfig or service account token

#### `namespace`
- **Type**: String
- **Required**: Yes
- **Description**: The Kubernetes namespace where pods will be created
- **Example**: `"nextflow"`, `"default"`, `"production-pipelines"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Namespace must exist in the cluster
  - Service account must have permissions in this namespace
  - Use dedicated namespaces for resource isolation

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: Storage location for Nextflow work directory
- **Format**: Depends on storage type:
  - S3: `s3://bucket-name/path`
  - GCS: `gs://bucket-name/path`
  - Azure Blob: `az://storage-account/container/path`
  - NFS/PVC: `/mnt/shared/work`
- **Examples**:
  - `"s3://my-nextflow-bucket/work"`
  - `"gs://my-bucket/work"`
  - `"az://mystorageaccount/work"`
  - `"/mnt/shared-storage/nextflow/work"`
- **Character Limit**: 0/400 characters
- **Notes**:
  - Must be accessible from all pods
  - Consider using persistent volumes or cloud storage
  - Ensure proper read/write permissions

### Optional Fields

#### `service_account`
- **Type**: String
- **Optional**: Yes
- **Description**: The Kubernetes service account for running workflow pods
- **Example**: `"nextflow-runner"`, `"default"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Service account must exist in the namespace
  - Must have RBAC permissions to create/manage pods
  - Different from credentials service account

### Compute Resources Block

#### `compute_resources`
Configuration for default compute resources for workflow tasks.

##### `compute_resources.cpus`
- **Type**: Integer or Float
- **Optional**: Yes
- **Description**: Number of CPUs allocated per task by default
- **Default**: `1`
- **Example**: `4`
- **Notes**:
  - Can be overridden per-process in Nextflow config
  - Maps to Kubernetes resource requests/limits

##### `compute_resources.memory`
- **Type**: String
- **Optional**: Yes
- **Description**: Memory allocation per task by default
- **Format**: Number with unit (GB, MB, Gi, Mi)
- **Default**: `"1 GB"`
- **Example**: `"16 GB"`, `"8192 Mi"`
- **Notes**:
  - Can be overridden per-process in Nextflow config
  - Maps to Kubernetes memory requests/limits

### Storage Claims Block

#### `storage_claims`
- **Type**: List of Objects
- **Optional**: Yes
- **Description**: Persistent Volume Claims (PVCs) to mount in workflow pods
- **Object Structure**:
  - `name` (String): Identifier for the storage claim
  - `mount_path` (String): Path where volume will be mounted
  - `claim_name` (String): Name of the PVC in Kubernetes
  - `read_only` (Boolean, optional): Mount as read-only
- **Example**:
  ```hcl
  storage_claims = [
    {
      name       = "shared-data"
      mount_path = "/mnt/data"
      claim_name = "shared-nfs-pvc"
      read_only  = false
    },
    {
      name       = "reference-genomes"
      mount_path = "/mnt/reference"
      claim_name = "reference-data-pvc"
      read_only  = true
    }
  ]
  ```
- **Notes**:
  - PVCs must exist in the namespace
  - Use for shared data, reference files, or output storage
  - Consider using ReadWriteMany (RWX) access mode for shared storage

### Pod Options Block

#### `pod_options`
Configuration for Kubernetes pod specifications.

##### `pod_options.node_selector`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Node selector labels to constrain pod scheduling to specific nodes
- **Example**:
  ```hcl
  node_selector = {
    "workload"     = "nextflow"
    "tier"         = "compute"
    "instance-type" = "n2-standard-8"
  }
  ```
- **Notes**:
  - Nodes must have matching labels
  - Useful for scheduling on specific node pools
  - Consider cost optimization (spot/preemptible nodes)

##### `pod_options.tolerations`
- **Type**: List of Objects
- **Optional**: Yes
- **Description**: Tolerations allow pods to schedule on tainted nodes
- **Object Structure**:
  - `key` (String): Taint key
  - `operator` (String): `"Equal"` or `"Exists"`
  - `value` (String, optional): Taint value (required if operator is Equal)
  - `effect` (String): `"NoSchedule"`, `"PreferNoSchedule"`, or `"NoExecute"`
- **Example**:
  ```hcl
  tolerations = [
    {
      key      = "dedicated"
      operator = "Equal"
      value    = "nextflow"
      effect   = "NoSchedule"
    },
    {
      key      = "spot"
      operator = "Exists"
      effect   = "NoSchedule"
    }
  ]
  ```
- **Notes**:
  - Required to schedule on tainted nodes
  - Useful for dedicated node pools
  - Common for spot/preemptible instances

##### `pod_options.labels`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Labels to apply to workflow pods
- **Example**:
  ```hcl
  labels = {
    "app"         = "nextflow"
    "managed-by"  = "seqera"
    "environment" = "production"
  }
  ```
- **Notes**:
  - Useful for organization and monitoring
  - Used by network policies and service selectors
  - Can be used for cost allocation

##### `pod_options.annotations`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Annotations to apply to workflow pods
- **Example**:
  ```hcl
  annotations = {
    "prometheus.io/scrape" = "true"
    "prometheus.io/port"   = "9090"
    "linkerd.io/inject"    = "enabled"
  }
  ```
- **Notes**:
  - Used for metadata and integrations
  - Common for monitoring (Prometheus), service mesh (Linkerd/Istio)
  - Not used for selection like labels

##### `pod_options.security_context`
- **Type**: Object
- **Optional**: Yes
- **Description**: Security context for pods
- **Object Structure**:
  - `run_as_user` (Integer): UID to run containers
  - `run_as_group` (Integer): GID to run containers
  - `fs_group` (Integer): Filesystem group for volumes
  - `run_as_non_root` (Boolean): Require non-root user
- **Example**:
  ```hcl
  security_context = {
    run_as_user     = 1000
    run_as_group    = 1000
    fs_group        = 1000
    run_as_non_root = true
  }
  ```
- **Notes**:
  - Important for security policies
  - Required by some PodSecurityPolicies/Standards

### Staging Options Block

#### `staging_options`
Configuration for workflow staging and lifecycle scripts.

##### `staging_options.pre_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script executed before pipeline launch
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Setting up Kubernetes environment..."
  export K8S_NAMESPACE=nextflow

  # Mount credentials or config
  kubectl get configmap my-config -o json

  # Validate storage access
  ls -la /mnt/shared-storage/
  ```
- **Use Cases**:
  - Validate Kubernetes resources
  - Check storage access
  - Set up environment
  - Download reference data

##### `staging_options.post_run_script`
- **Type**: String
- **Optional**: Yes
- **Description**: Bash script executed after pipeline completion
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Pipeline completed with exit code: $NXF_EXIT_STATUS"

  # Archive results
  kubectl cp results/ result-archive:/archive/

  # Cleanup
  rm -rf /tmp/work/*
  ```
- **Use Cases**:
  - Archive results
  - Cleanup resources
  - Send notifications
  - Generate reports

### Nextflow Configuration

#### `nextflow_config`
- **Type**: String
- **Optional**: Yes
- **Description**: Global Nextflow configuration for Kubernetes executor
- **Format**: Nextflow configuration DSL
- **Character Limit**: 0/3200 characters
- **Example**:
  ```groovy
  process {
    executor = 'k8s'
    container = 'nextflow/nextflow:23.10.0'

    errorStrategy = 'retry'
    maxRetries = 3

    cpus = 2
    memory = '4 GB'

    withLabel: big_mem {
      memory = '32 GB'
      cpus = 8
    }
  }

  k8s {
    namespace = 'nextflow'
    serviceAccount = 'nextflow-runner'

    pod {
      nodeSelector = 'workload=nextflow'

      securityContext {
        runAsUser = 1000
        fsGroup = 1000
      }
    }
  }

  docker {
    enabled = true
    runOptions = '-u $(id -u):$(id -g)'
  }
  ```
- **Use Cases**:
  - Configure Kubernetes-specific options
  - Set resource requirements
  - Define pod specifications
  - Configure node selection

### Environment Variables

#### `environment_variables`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables set in all workflow pods
- **Example**:
  ```hcl
  environment_variables = {
    "NXF_ANSI_LOG"    = "false"
    "NXF_OPTS"        = "-Xms1g -Xmx4g"
    "K8S_NAMESPACE"   = "nextflow"
    "TMPDIR"          = "/tmp"
  }
  ```
- **Notes**:
  - Available to all processes
  - Can reference Kubernetes ConfigMaps/Secrets via Nextflow config

### Advanced Options Block

#### `advanced`
Advanced configuration options for Kubernetes integration.

##### `advanced.head_service_account`
- **Type**: String
- **Optional**: Yes
- **Description**: Service account for the head/launcher pod (different from task pods)
- **Example**: `"nextflow-head"`, `"nextflow-launcher"`
- **Notes**:
  - Used for the Nextflow orchestrator pod
  - May need additional permissions (create/list/delete pods)
  - Falls back to `service_account` if not specified

##### `advanced.image_pull_policy`
- **Type**: String
- **Optional**: Yes
- **Description**: Image pull policy for container images
- **Allowed Values**: `"Always"`, `"IfNotPresent"`, `"Never"`
- **Default**: `"IfNotPresent"`
- **Example**: `"Always"`
- **Notes**:
  - `Always`: Pull image on every pod start
  - `IfNotPresent`: Pull only if not cached locally
  - `Never`: Never pull, must be present locally

##### `advanced.image_pull_secrets`
- **Type**: List of Strings
- **Optional**: Yes
- **Description**: Names of Kubernetes secrets for pulling private container images
- **Example**: `["docker-registry-secret", "gcr-secret"]`
- **Notes**:
  - Secrets must exist in the namespace
  - Secrets must be type `kubernetes.io/dockerconfigjson`
  - Required for private container registries

##### `advanced.storage_class_name`
- **Type**: String
- **Optional**: Yes
- **Description**: StorageClass name for dynamic volume provisioning
- **Example**: `"fast-ssd"`, `"standard"`, `"nfs-client"`
- **Notes**:
  - Used for dynamically provisioned PVCs
  - StorageClass must exist in the cluster
  - Affects performance and cost

##### `advanced.storage_mount_path`
- **Type**: String
- **Optional**: Yes
- **Description**: Default mount path for storage volumes
- **Default**: `"/workspace"`
- **Example**: `"/mnt/data"`
- **Notes**:
  - Used as base path for mounted volumes
  - Can be overridden per storage claim

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

1. **namespace**: Must be valid Kubernetes namespace name
2. **service_account**: Must exist in the namespace
3. **work_directory**: Must be accessible from pods
4. **storage_claims**: PVC names must exist in namespace
5. **pod_options.tolerations**: Must have valid operator and effect values
6. **advanced.image_pull_policy**: Must be one of: Always, IfNotPresent, Never

### Lifecycle Considerations

- **Create**: Configures Kubernetes compute environment in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields
- **Delete**: Removes compute environment (doesn't delete Kubernetes resources)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `credentials_id`
- `namespace`

### Mutable Fields

These fields can be updated without replacement:
- `service_account`
- `work_directory`
- `compute_resources`
- `storage_claims`
- `pod_options`
- `staging_options`
- `nextflow_config`
- `environment_variables`
- `advanced` options

### Sensitive Fields

- The referenced `credentials_id` points to sensitive Kubernetes credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration

```hcl
resource "seqera_compute_kubernetes" "minimal" {
  name           = "k8s-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.main.credentials_id
  namespace      = "nextflow"
  work_directory = "s3://my-bucket/work"
}
```

### Standard Configuration with Storage

```hcl
resource "seqera_compute_kubernetes" "standard" {
  name           = "k8s-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.main.credentials_id
  namespace      = "nextflow"
  service_account = "nextflow-runner"
  work_directory = "s3://my-bucket/work"

  compute_resources {
    cpus   = 4
    memory = "16 GB"
  }

  storage_claims = [
    {
      name       = "shared-data"
      mount_path = "/mnt/data"
      claim_name = "shared-nfs-pvc"
    }
  ]
}
```

### Production with Node Selection and Tolerations

```hcl
resource "seqera_compute_kubernetes" "production" {
  name           = "k8s-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.main.credentials_id
  namespace      = "production-pipelines"
  service_account = "nextflow-prod"
  work_directory = "s3://prod-bucket/work"

  compute_resources {
    cpus   = 8
    memory = "32 GB"
  }

  storage_claims = [
    {
      name       = "shared-work"
      mount_path = "/mnt/shared"
      claim_name = "prod-shared-pvc"
    },
    {
      name       = "reference-data"
      mount_path = "/mnt/reference"
      claim_name = "reference-pvc"
      read_only  = true
    }
  ]

  pod_options {
    node_selector = {
      "workload"      = "nextflow"
      "tier"          = "compute"
      "instance-type" = "n2-highmem-8"
    }

    tolerations = [
      {
        key      = "dedicated"
        operator = "Equal"
        value    = "nextflow"
        effect   = "NoSchedule"
      }
    ]

    labels = {
      "app"         = "nextflow"
      "environment" = "production"
      "managed-by"  = "seqera"
    }

    annotations = {
      "prometheus.io/scrape" = "true"
      "prometheus.io/port"   = "9090"
    }

    security_context = {
      run_as_user     = 1000
      run_as_group    = 1000
      fs_group        = 1000
      run_as_non_root = true
    }
  }

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Validating production environment..."

      # Check namespace
      kubectl get namespace production-pipelines

      # Validate storage access
      ls -la /mnt/shared
      ls -la /mnt/reference

      echo "Environment validated"
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed: $NXF_EXIT_STATUS"

      # Archive results
      aws s3 sync /tmp/results s3://archive-bucket/$(date +%Y-%m-%d)/

      # Cleanup
      rm -rf /tmp/work/*
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'k8s'
      container = 'nextflow/nextflow:23.10.0'

      errorStrategy = 'retry'
      maxRetries = 2

      cpus = 8
      memory = '32 GB'

      withLabel: intensive {
        cpus = 16
        memory = '64 GB'
      }
    }

    k8s {
      namespace = 'production-pipelines'
      serviceAccount = 'nextflow-prod'

      pod {
        nodeSelector = 'workload=nextflow'

        securityContext {
          runAsUser = 1000
          fsGroup = 1000
        }
      }
    }
  EOF

  environment_variables = {
    "NXF_ANSI_LOG"    = "false"
    "NXF_OPTS"        = "-Xms2g -Xmx8g"
    "AWS_REGION"      = "us-east-1"
    "K8S_NAMESPACE"   = "production-pipelines"
  }

  advanced {
    head_service_account = "nextflow-head"
    image_pull_policy    = "IfNotPresent"
    image_pull_secrets   = ["ecr-registry-secret"]
  }
}
```

### GKE Configuration with Spot Nodes

```hcl
resource "seqera_compute_kubernetes" "gke_spot" {
  name           = "k8s-gke-spot"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.gke.credentials_id
  namespace      = "nextflow"
  work_directory = "gs://my-bucket/work"

  compute_resources {
    cpus   = 4
    memory = "16 GB"
  }

  pod_options {
    node_selector = {
      "cloud.google.com/gke-spot" = "true"
      "workload"                  = "nextflow"
    }

    tolerations = [
      {
        key      = "cloud.google.com/gke-spot"
        operator = "Equal"
        value    = "true"
        effect   = "NoSchedule"
      }
    ]

    labels = {
      "app"      = "nextflow"
      "platform" = "gke"
    }
  }

  nextflow_config = <<-EOF
    process {
      errorStrategy = 'retry'
      maxRetries = 3
      maxErrors = -1
    }
  EOF

  environment_variables = {
    "GOOGLE_CLOUD_PROJECT" = var.gcp_project_id
  }
}
```

### EKS Configuration with Fargate

```hcl
resource "seqera_compute_kubernetes" "eks_fargate" {
  name           = "k8s-eks-fargate"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.eks.credentials_id
  namespace      = "fargate-nextflow"
  work_directory = "s3://my-bucket/work"

  compute_resources {
    cpus   = 2
    memory = "8 GB"
  }

  pod_options {
    labels = {
      "app"      = "nextflow"
      "platform" = "eks-fargate"
    }

    annotations = {
      "eks.amazonaws.com/fargate-profile" = "nextflow"
    }
  }

  environment_variables = {
    "AWS_REGION" = "us-east-1"
  }

  advanced {
    head_service_account = "nextflow-fargate"
    image_pull_policy    = "Always"
  }
}
```

### AKS Configuration with Azure Storage

```hcl
resource "seqera_compute_kubernetes" "aks" {
  name           = "k8s-aks"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.aks.credentials_id
  namespace      = "nextflow"
  work_directory = "az://mystorageaccount/work"

  compute_resources {
    cpus   = 4
    memory = "16 GB"
  }

  storage_claims = [
    {
      name       = "azure-files"
      mount_path = "/mnt/azurefiles"
      claim_name = "azure-files-pvc"
    }
  ]

  pod_options {
    node_selector = {
      "agentpool" = "nextflow"
    }

    labels = {
      "app"      = "nextflow"
      "platform" = "aks"
    }
  }

  environment_variables = {
    "AZURE_STORAGE_ACCOUNT" = "mystorageaccount"
  }

  advanced {
    storage_class_name = "azure-file-premium"
  }
}
```

### GPU-Enabled Configuration

```hcl
resource "seqera_compute_kubernetes" "gpu" {
  name           = "k8s-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.main.credentials_id
  namespace      = "gpu-pipelines"
  work_directory = "s3://gpu-bucket/work"

  compute_resources {
    cpus   = 8
    memory = "32 GB"
  }

  pod_options {
    node_selector = {
      "accelerator" = "nvidia-tesla-v100"
    }

    tolerations = [
      {
        key      = "nvidia.com/gpu"
        operator = "Exists"
        effect   = "NoSchedule"
      }
    ]

    labels = {
      "gpu" = "true"
    }
  }

  nextflow_config = <<-EOF
    process {
      withLabel: gpu {
        accelerator = 1
        containerOptions = '--gpus all'
      }
    }
  EOF

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0"
  }
}
```

### High-Security Configuration

```hcl
resource "seqera_compute_kubernetes" "secure" {
  name           = "k8s-secure"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.main.credentials_id
  namespace      = "secure-pipelines"
  service_account = "nextflow-secure"
  work_directory = "s3://secure-bucket/work"

  pod_options {
    security_context = {
      run_as_user     = 1000
      run_as_group    = 1000
      fs_group        = 1000
      run_as_non_root = true
    }

    labels = {
      "security-level" = "high"
    }

    annotations = {
      "container.apparmor.security.beta.kubernetes.io/nextflow" = "runtime/default"
      "seccomp.security.alpha.kubernetes.io/pod"                = "runtime/default"
    }
  }

  advanced {
    image_pull_policy  = "Always"
    image_pull_secrets = ["secure-registry-secret"]
  }

  environment_variables = {
    "SECURITY_LEVEL" = "high"
  }
}
```

## Integration with Terraform Kubernetes Provider

### Complete Example with Infrastructure

```hcl
# Kubernetes Provider Configuration
provider "kubernetes" {
  config_path = "~/.kube/config"
}

# Namespace
resource "kubernetes_namespace" "nextflow" {
  metadata {
    name = "nextflow"

    labels = {
      name        = "nextflow"
      environment = "production"
    }
  }
}

# Service Account for tasks
resource "kubernetes_service_account" "nextflow_runner" {
  metadata {
    name      = "nextflow-runner"
    namespace = kubernetes_namespace.nextflow.metadata[0].name

    annotations = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.nextflow.arn  # For IRSA on EKS
    }
  }
}

# Service Account for head pod
resource "kubernetes_service_account" "nextflow_head" {
  metadata {
    name      = "nextflow-head"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }
}

# Role for task pods
resource "kubernetes_role" "nextflow_runner" {
  metadata {
    name      = "nextflow-runner"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  rule {
    api_groups = [""]
    resources  = ["pods", "pods/log", "pods/status"]
    verbs      = ["get", "list", "watch"]
  }

  rule {
    api_groups = [""]
    resources  = ["pods/exec"]
    verbs      = ["create"]
  }
}

# Role for head pod
resource "kubernetes_role" "nextflow_head" {
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

# RoleBinding for runner
resource "kubernetes_role_binding" "nextflow_runner" {
  metadata {
    name      = "nextflow-runner"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.nextflow_runner.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.nextflow_runner.metadata[0].name
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }
}

# RoleBinding for head
resource "kubernetes_role_binding" "nextflow_head" {
  metadata {
    name      = "nextflow-head"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.nextflow_head.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.nextflow_head.metadata[0].name
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }
}

# Persistent Volume Claim for shared storage
resource "kubernetes_persistent_volume_claim" "shared" {
  metadata {
    name      = "shared-nfs-pvc"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteMany"]

    resources {
      requests = {
        storage = "100Gi"
      }
    }

    storage_class_name = "nfs-client"
  }
}

# Image Pull Secret for private registry
resource "kubernetes_secret" "docker_registry" {
  metadata {
    name      = "docker-registry-secret"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  type = "kubernetes.io/dockerconfigjson"

  data = {
    ".dockerconfigjson" = jsonencode({
      auths = {
        "https://index.docker.io/v1/" = {
          username = var.docker_username
          password = var.docker_password
          auth     = base64encode("${var.docker_username}:${var.docker_password}")
        }
      }
    })
  }
}

# ConfigMap for shared configuration
resource "kubernetes_config_map" "nextflow_config" {
  metadata {
    name      = "nextflow-config"
    namespace = kubernetes_namespace.nextflow.metadata[0].name
  }

  data = {
    "nextflow.config" = <<-EOF
      process {
        executor = 'k8s'
        container = 'nextflow/nextflow:latest'
      }
    EOF
  }
}

# Seqera Compute Environment
resource "seqera_compute_kubernetes" "integrated" {
  name           = "k8s-integrated"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_kubernetes_credential.main.credentials_id
  namespace      = kubernetes_namespace.nextflow.metadata[0].name
  service_account = kubernetes_service_account.nextflow_runner.metadata[0].name
  work_directory = "s3://my-bucket/work"

  compute_resources {
    cpus   = 4
    memory = "16 GB"
  }

  storage_claims = [
    {
      name       = "shared-storage"
      mount_path = "/mnt/shared"
      claim_name = kubernetes_persistent_volume_claim.shared.metadata[0].name
    }
  ]

  pod_options {
    labels = {
      "app"        = "nextflow"
      "managed-by" = "terraform"
    }
  }

  advanced {
    head_service_account = kubernetes_service_account.nextflow_head.metadata[0].name
    image_pull_secrets   = [kubernetes_secret.docker_registry.metadata[0].name]
  }

  depends_on = [
    kubernetes_role_binding.nextflow_runner,
    kubernetes_role_binding.nextflow_head,
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
- Resource `namespace` → API `config.namespace`
- Resource `service_account` → API `config.serviceAccount`
- Resource `work_directory` → API `config.workDir`
- Resource `compute_resources` → API `config.cpus`, `config.memory`
- Resource `storage_claims` → API `config.storageClaims`
- Resource `pod_options.node_selector` → API `config.nodeSelector`
- Resource `pod_options.tolerations` → API `config.tolerations`
- Resource `pod_options.labels` → API `config.podLabels`
- Resource `pod_options.annotations` → API `config.podAnnotations`
- Resource `advanced.head_service_account` → API `config.headServiceAccount`
- Resource `advanced.image_pull_policy` → API `config.imagePullPolicy`
- Resource `advanced.image_pull_secrets` → API `config.imagePullSecrets`

### Platform Type

- **API Value**: `"k8s"` or `"kubernetes"`
- **Config Type**: `"K8sComputeConfig"` or `"KubernetesComputeConfig"`

## Related Resources

- `seqera_kubernetes_credential` - Kubernetes credentials (kubeconfig)
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Kubernetes Cluster Requirements

### Required RBAC Permissions

The service account specified in credentials must have:

#### For Head/Launcher Pod
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: nextflow-head
rules:
- apiGroups: [""]
  resources: ["pods", "pods/status", "pods/log"]
  verbs: ["get", "list", "watch", "create", "delete"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "create", "delete"]
```

#### For Task Pods
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: nextflow-runner
rules:
- apiGroups: [""]
  resources: ["pods", "pods/status", "pods/log"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create"]
```

### Cluster Configuration

- **Kubernetes Version**: 1.20+
- **Container Runtime**: Docker, containerd, or CRI-O
- **Storage**: PersistentVolumes with appropriate access modes
- **Networking**: Pod-to-pod communication enabled
- **Resource Quotas**: Sufficient CPU and memory quotas

### Storage Options

1. **Cloud Storage** (Recommended):
   - AWS: S3 with IAM roles (IRSA on EKS)
   - GCP: GCS with Workload Identity (on GKE)
   - Azure: Blob Storage with Managed Identity (on AKS)

2. **Persistent Volumes**:
   - NFS with ReadWriteMany (RWX)
   - CephFS or GlusterFS
   - Cloud provider file storage (EFS, Azure Files, Filestore)

3. **Local Storage**:
   - HostPath (not recommended for production)
   - Local PersistentVolumes

## Best Practices

### Resource Management

1. **Set Resource Limits**: Define CPU and memory limits to prevent resource exhaustion
2. **Use Node Selectors**: Direct workloads to appropriate node pools
3. **Implement Tolerations**: Schedule on tainted nodes (spot instances)
4. **Resource Quotas**: Set namespace quotas to prevent over-provisioning
5. **Limit Ranges**: Define default and max resources per pod

### Storage

1. **Use Cloud Storage**: S3/GCS/Azure Blob for work directory
2. **ReadWriteMany**: Use RWX PVCs for shared data
3. **Storage Classes**: Use appropriate storage classes for performance/cost
4. **Cleanup**: Implement lifecycle policies for work directory
5. **Separate Volumes**: Use different PVCs for work, reference data, and results

### Security

1. **Service Accounts**: Use dedicated service accounts with minimal permissions
2. **RBAC**: Implement principle of least privilege
3. **Pod Security**: Use SecurityContext and PodSecurityPolicies/Standards
4. **Network Policies**: Restrict pod-to-pod communication
5. **Image Pull Secrets**: Secure private registry access
6. **Secrets Management**: Use Kubernetes Secrets or external secret managers

### Performance

1. **Node Affinity**: Co-locate related pods
2. **Horizontal Scaling**: Use HPA for autoscaling
3. **Resource Requests**: Set appropriate requests for scheduling
4. **Fast Storage**: Use SSD-backed storage for work directory
5. **Network**: Ensure high-bandwidth networking for data-intensive workflows

### Reliability

1. **Error Handling**: Configure retry strategies in Nextflow
2. **Health Checks**: Implement liveness and readiness probes
3. **Monitoring**: Use Prometheus for metrics collection
4. **Logging**: Aggregate logs with ELK/Loki
5. **Backup**: Regularly backup important data and configurations

### Cost Optimization

1. **Spot Instances**: Use spot/preemptible nodes with tolerations
2. **Cluster Autoscaler**: Scale nodes based on workload
3. **Resource Quotas**: Prevent over-provisioning
4. **Storage Cleanup**: Delete old work files automatically
5. **Right-Sizing**: Use appropriate instance types and resource requests

## Troubleshooting

### Common Issues

1. **Pod Fails to Schedule**:
   - Check node selectors and tolerations
   - Verify resource availability
   - Check PVC binding status

2. **Permission Denied**:
   - Verify service account RBAC permissions
   - Check SecurityContext UID/GID
   - Verify storage access permissions

3. **Image Pull Errors**:
   - Verify image pull secrets
   - Check image repository access
   - Verify image pull policy

4. **Storage Issues**:
   - Check PVC status (Bound/Pending)
   - Verify StorageClass availability
   - Check volume mount permissions

5. **Network Problems**:
   - Verify network policies
   - Check service discovery
   - Verify DNS resolution

## References

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Nextflow Kubernetes Executor](https://www.nextflow.io/docs/latest/kubernetes.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
