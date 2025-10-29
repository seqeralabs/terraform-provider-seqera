# Amazon EKS Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_amazon_eks` resource, which manages Amazon EKS (Elastic Kubernetes Service) compute environments in Seqera Platform.

Amazon EKS compute environments enable running Nextflow workflows on AWS-managed Kubernetes clusters, combining the benefits of Kubernetes container orchestration with AWS-native integrations.

## Key Characteristics

- **AWS-Managed Kubernetes**: Fully managed control plane by AWS
- **AWS Integration**: Native S3, IAM, VPC, and CloudWatch integration
- **Fusion Storage**: Optional high-performance S3 access via Fusion v2
- **IRSA Support**: IAM Roles for Service Accounts for pod-level permissions
- **Fargate Support**: Serverless compute with AWS Fargate
- **Scalability**: Cluster Autoscaler and Karpenter support

## Resource Structure

```hcl
resource "seqera_compute_amazon_eks" "example" {
  name         = "eks-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"

  # EKS cluster
  cluster_name = "my-eks-cluster"
  namespace    = "nextflow"

  # Storage mode
  storage_mode = "fusion"  # or "legacy"

  # Work directory (S3)
  work_directory = "s3://my-bucket/work"

  # Service accounts
  head_service_account    = "tower-launcher-sa"
  compute_service_account = "tower-job-sa"

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
    "AWS_REGION" = "us-east-1"
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
- **Example**: `"eks-prod"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: AWS credentials ID to use for accessing EKS and AWS services
- **Reference**: Must reference a valid `seqera_aws_credential` resource
- **Example**: `seqera_aws_credential.main.credentials_id`
- **Notes**: Credentials must have permissions for:
  - EKS cluster access
  - S3 bucket access
  - IAM (for IRSA)
  - EC2 (for node management)

#### `region`
- **Type**: String
- **Required**: Yes
- **Description**: The target execution region (AWS region where the EKS cluster is located)
- **Validation**: Must be a valid AWS region code
- **Examples**: `"us-east-1"`, `"us-west-2"`, `"eu-west-1"`, `"ap-southeast-1"`
- **Character Limit**: 0/100 characters
- **Notes**: EKS cluster must be in this region

#### `cluster_name`
- **Type**: String
- **Required**: Yes
- **Description**: The AWS EKS cluster name
- **Example**: `"my-eks-cluster"`, `"production-eks"`
- **Character Limit**: 0/100 characters
- **Validation**: Error shown if empty - "Value is required"
- **Notes**:
  - Cluster must exist in the specified region
  - Credentials must have access to the cluster
  - Can be obtained from AWS EKS console or CLI

#### `namespace`
- **Type**: String
- **Required**: Yes
- **Description**: The Kubernetes namespace to use for the pipeline execution
- **Example**: `"nextflow"`, `"tower-nf"`, `"default"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Namespace must exist in the EKS cluster
  - Service accounts must have permissions in this namespace

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: The S3 bucket path to be used as pipeline work directory
- **Format**: `s3://bucket-name/path`
- **Example**: `"s3://my-nextflow-bucket/work"`
- **Character Limit**: 0/200 characters
- **Constraints**:
  - Must start with `s3://`
  - **The S3 bucket must be located in the same region entered above**
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
- **Description**: Allow access to your AWS S3-hosted data via the Fusion v2 virtual distributed file system, speeding up most operations
- **Requirements**: Requires configuring a Shared file system in your AWS/Kubernetes cluster
- **Benefits**:
  - High-performance S3 access
  - Lazy loading of files
  - Reduced data transfer costs
  - Improved pipeline performance
- **Notes**: This enables Fusion v2 for optimized S3 access

##### Legacy Storage Mode
- **Description**: For this option, the Nextflow work directory for your data pipeline must be located on a POSIX-compatible file system
- **Requirements**: You must configure a shared file system in your Kubernetes cluster
- **Notes**:
  - Requires NFS, EFS, or other POSIX-compatible shared storage
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
  - Can use IRSA (IAM Roles for Service Accounts) for AWS permissions

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
  - Can use IRSA for S3 access without AWS credentials
  - Default is "default" service account

### Resource Labels

#### `resource_labels`
- **Type**: List of Objects
- **Optional**: Yes
- **Description**: Associate name/value pairs with the resources created by this compute environment
- **Object Structure**:
  - `name` (String): Label/tag key
  - `value` (String): Label/tag value
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
  - Labels applied to Kubernetes pods and AWS resources

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
  echo "Setting up EKS environment..."
  export AWS_REGION=us-east-1

  # Verify cluster access
  kubectl get nodes

  # Check S3 access
  aws s3 ls s3://my-bucket/

  # Download reference data
  aws s3 sync s3://reference-bucket/genome /mnt/reference/
  ```
- **Use Cases**:
  - Validate EKS cluster connectivity
  - Check S3 access and permissions
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

  # Archive results to S3
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  aws s3 sync /tmp/results s3://archive-bucket/results-$TIMESTAMP/

  # Cleanup
  rm -rf /tmp/work/*
  ```
- **Use Cases**:
  - Archive results to S3
  - Send notifications via SNS
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

  aws {
    region = 'us-east-1'
    batch.cliPath = '/home/ec2-user/miniconda/bin/aws'
  }
  ```
- **Use Cases**:
  - Configure Kubernetes executor
  - Set AWS-specific options
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
    "AWS_REGION"           = "us-east-1"
    "NXF_ANSI_LOG"         = "false"
    "NXF_OPTS"             = "-Xms1g -Xmx4g"
    "FUSION_ENABLED"       = "true"
  }
  ```
- **Notes**:
  - Available to all processes
  - AWS credentials automatically available via IRSA if configured

### Advanced Options Block

#### `advanced`
Advanced configuration options for EKS integration.

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
      instance-type: m5.xlarge
    tolerations:
    - key: dedicated
      operator: Equal
      value: nextflow
      effect: NoSchedule
    affinity:
      podAntiAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nextflow
            topologyKey: kubernetes.io/hostname
    securityContext:
      runAsUser: 1000
      fsGroup: 1000
  ```
- **Notes**:
  - Applies to the Nextflow head/launcher pod
  - Can specify nodeSelector, affinity, tolerations, securityContext
  - Must be valid Kubernetes PodSpec YAML
  - Allows for Fargate scheduling, spot instances, etc.

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

1. **region**: Must be a valid AWS region
2. **cluster_name**: Required, cannot be empty
3. **namespace**: Must be valid Kubernetes namespace
4. **work_directory**: Must start with `s3://` and be in the same region
5. **storage_mode**: Must be either "fusion" or "legacy"
6. **pod_cleanup_policy**: Must be one of: on_success, always, never
7. **custom_head_pod_specs**: Must be valid PodSpec YAML starting with "spec:"

### Lifecycle Considerations

- **Create**: Configures EKS compute environment in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields
- **Delete**: Removes compute environment (doesn't delete EKS cluster)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `credentials_id`
- `region`
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

- The referenced `credentials_id` points to sensitive AWS credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration with Fusion

```hcl
resource "seqera_compute_amazon_eks" "minimal" {
  name           = "eks-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = "my-eks-cluster"
  namespace      = "nextflow"
  storage_mode   = "fusion"
  work_directory = "s3://my-bucket/work"
}
```

### Standard Configuration with Service Accounts

```hcl
resource "seqera_compute_amazon_eks" "standard" {
  name           = "eks-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = "production-eks"
  namespace      = "nextflow"
  storage_mode   = "fusion"
  work_directory = "s3://prod-bucket/work"

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

### Legacy Storage with EFS

```hcl
resource "seqera_compute_amazon_eks" "legacy" {
  name           = "eks-legacy-efs"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = "my-eks-cluster"
  namespace      = "nextflow"
  storage_mode   = "legacy"
  work_directory = "/mnt/efs/nextflow-work"

  # Note: Requires EFS CSI driver and PVC configured in cluster
  resource_labels = [
    {
      name  = "storage_type"
      value = "efs"
    }
  ]
}
```

### Production with Custom Head Pod Specs

```hcl
resource "seqera_compute_amazon_eks" "production" {
  name           = "eks-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = "production-eks"
  namespace      = "production-pipelines"
  storage_mode   = "fusion"
  work_directory = "s3://prod-bucket/work"

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
      echo "Initializing production EKS environment..."

      # Verify cluster access
      kubectl get nodes --no-headers | wc -l

      # Check S3 access
      aws s3 ls s3://prod-bucket/ || exit 1

      # Download reference data
      aws s3 sync s3://reference-bucket/genome /mnt/reference/

      echo "Environment ready"
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed: $NXF_EXIT_STATUS"

      # Archive results
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      aws s3 sync /tmp/results s3://archive-bucket/results-$TIMESTAMP/

      # Send SNS notification
      aws sns publish \
        --topic-arn arn:aws:sns:us-east-1:123456789012:pipeline-complete \
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

    aws {
      region = 'us-east-1'
    }
  EOF

  environment_variables = {
    "AWS_REGION"    = "us-east-1"
    "NXF_ANSI_LOG"  = "false"
    "NXF_OPTS"      = "-Xms2g -Xmx8g"
    "FUSION_ENABLED" = "true"
  }

  advanced {
    pod_cleanup_policy = "on_success"

    custom_head_pod_specs = <<-YAML
      spec:
        nodeSelector:
          workload: nextflow
          instance-type: m5.xlarge
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

### Spot Instances Configuration

```hcl
resource "seqera_compute_amazon_eks" "spot" {
  name           = "eks-spot"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = "my-eks-cluster"
  namespace      = "nextflow"
  storage_mode   = "fusion"
  work_directory = "s3://my-bucket/work"

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
        nodeSelector = 'eks.amazonaws.com/capacityType=SPOT'

        tolerations = [[
          key: 'spot',
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
          eks.amazonaws.com/capacityType: SPOT
        tolerations:
        - key: spot
          operator: Equal
          value: "true"
          effect: NoSchedule
    YAML
  }
}
```

### Fargate Configuration

```hcl
resource "seqera_compute_amazon_eks" "fargate" {
  name           = "eks-fargate"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = "fargate-eks-cluster"
  namespace      = "fargate-nextflow"
  storage_mode   = "fusion"
  work_directory = "s3://fargate-bucket/work"

  head_service_account    = "fargate-launcher"
  compute_service_account = "fargate-job"

  resource_labels = [
    {
      name  = "compute_type"
      value = "fargate"
    }
  ]

  environment_variables = {
    "AWS_REGION" = "us-east-1"
  }

  advanced {
    # Fargate profile will automatically schedule pods
    # No custom pod specs needed if Fargate profile matches namespace
    head_job_cpus   = 2
    head_job_memory = 4096
  }
}
```

### GPU-Enabled Configuration

```hcl
resource "seqera_compute_amazon_eks" "gpu" {
  name           = "eks-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = "gpu-eks-cluster"
  namespace      = "gpu-pipelines"
  storage_mode   = "fusion"
  work_directory = "s3://gpu-bucket/work"

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
        nodeSelector = 'node.kubernetes.io/instance-type=p3.2xlarge'
      }
    }
  EOF

  advanced {
    custom_head_pod_specs = <<-YAML
      spec:
        nodeSelector:
          node.kubernetes.io/instance-type: m5.xlarge
    YAML
  }

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0"
  }
}
```

## Integration with Terraform AWS Provider

### Complete Example with EKS Infrastructure

```hcl
# VPC for EKS
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "nextflow-eks-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b", "us-east-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true

  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    "kubernetes.io/cluster/nextflow-eks" = "shared"
  }
}

# EKS Cluster
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "nextflow-eks"
  cluster_version = "1.28"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access = true

  eks_managed_node_groups = {
    nextflow = {
      min_size     = 1
      max_size     = 10
      desired_size = 2

      instance_types = ["m5.xlarge"]

      labels = {
        workload = "nextflow"
      }

      taints = [{
        key    = "dedicated"
        value  = "nextflow"
        effect = "NoSchedule"
      }]
    }
  }

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

# S3 Bucket for work directory
resource "aws_s3_bucket" "work" {
  bucket = "nextflow-eks-work-bucket"

  tags = {
    Purpose = "nextflow-work"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "work" {
  bucket = aws_s3_bucket.work.id

  rule {
    id     = "delete-old-work-files"
    status = "Enabled"

    expiration {
      days = 30
    }
  }
}

# IAM Role for Service Account (IRSA) - Head Pod
module "irsa_head" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "nextflow-head-irsa"

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["nextflow:tower-launcher-sa"]
    }
  }

  role_policy_arns = {
    s3_access = aws_iam_policy.s3_access.arn
  }
}

# IAM Role for Service Account (IRSA) - Worker Pods
module "irsa_worker" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "nextflow-worker-irsa"

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["nextflow:tower-job-sa"]
    }
  }

  role_policy_arns = {
    s3_access = aws_iam_policy.s3_access.arn
  }
}

# IAM Policy for S3 Access
resource "aws_iam_policy" "s3_access" {
  name = "nextflow-s3-access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.work.arn,
          "${aws_s3_bucket.work.arn}/*"
        ]
      }
    ]
  })
}

# Kubernetes Provider
provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  token                  = data.aws_eks_cluster_auth.this.token
}

data "aws_eks_cluster_auth" "this" {
  name = module.eks.cluster_name
}

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
      "eks.amazonaws.com/role-arn" = module.irsa_head.iam_role_arn
    }
  }
}

# Service Account for Worker Pods
resource "kubernetes_service_account" "worker" {
  metadata {
    name      = "tower-job-sa"
    namespace = kubernetes_namespace.nextflow.metadata[0].name

    annotations = {
      "eks.amazonaws.com/role-arn" = module.irsa_worker.iam_role_arn
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
resource "seqera_compute_amazon_eks" "integrated" {
  name           = "eks-integrated"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  cluster_name   = module.eks.cluster_name
  namespace      = kubernetes_namespace.nextflow.metadata[0].name
  storage_mode   = "fusion"
  work_directory = "s3://${aws_s3_bucket.work.id}/work"

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
- Resource `region` → API `config.region`
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

- **API Value**: `"eks"` or `"amazon-eks"`
- **Config Type**: `"EKSComputeConfig"` or `"AmazonEKSComputeConfig"`

## Related Resources

- `seqera_aws_credential` - AWS credentials for EKS and S3 access
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## EKS Cluster Requirements

### Required AWS Permissions

The AWS credentials must have:
- `eks:DescribeCluster`
- `eks:ListClusters`
- `s3:GetObject`, `s3:PutObject`, `s3:ListBucket` (for work bucket)
- `iam:GetRole` (if using IRSA)

### Required Kubernetes RBAC

Service accounts need appropriate RBAC permissions (see examples above).

### Cluster Configuration

- **Kubernetes Version**: 1.20+
- **OIDC Provider**: Enabled for IRSA
- **VPC**: Private subnets for nodes
- **Storage**: S3 for Fusion mode, EFS/FSx for Legacy mode
- **Node Groups**: Appropriate instance types and autoscaling

## Storage Options Comparison

| Feature | Fusion Storage | Legacy Storage |
|---------|----------------|----------------|
| **Backend** | S3 via Fusion v2 | POSIX filesystem (EFS/FSx) |
| **Performance** | High (lazy loading) | Depends on filesystem |
| **Setup** | Simpler (just S3) | Requires EFS CSI driver + PVC |
| **Cost** | S3 storage costs | EFS/FSx costs (higher) |
| **Scalability** | S3 unlimited | Limited by filesystem |
| **Best For** | Most workflows | POSIX-requiring workflows |

## Best Practices

### IRSA (IAM Roles for Service Accounts)

1. **Use IRSA**: Avoid AWS credentials in pods
2. **Separate Roles**: Different roles for head and worker pods
3. **Least Privilege**: Only grant necessary S3 permissions
4. **Multiple Buckets**: Different roles for different buckets

### Node Selection

1. **Node Selectors**: Direct workloads to appropriate nodes
2. **Spot Instances**: Use with tolerations for cost savings
3. **Fargate**: Serverless option for variable workloads
4. **GPU Nodes**: Dedicated node groups for GPU workloads

### Storage

1. **Use Fusion**: Recommended for most workflows
2. **Regional S3**: Bucket in same region as EKS
3. **Lifecycle Policies**: Delete old work files automatically
4. **Versioning**: Enable for important results buckets

### Security

1. **Private Cluster**: Use private endpoint for EKS API
2. **VPC**: Deploy in private subnets
3. **Security Groups**: Restrict network access
4. **Pod Security**: Use SecurityContext and Pod Security Standards
5. **Secrets**: Use AWS Secrets Manager or Kubernetes Secrets

### Performance

1. **Fusion v2**: Faster than legacy EFS/FSx
2. **Instance Types**: Choose appropriate compute-optimized types
3. **Cluster Autoscaler**: Auto-scale nodes based on workload
4. **Network**: Ensure high bandwidth between pods and S3

### Cost Optimization

1. **Spot Instances**: 70% cheaper with proper tolerations
2. **Fargate**: Pay only for pod runtime
3. **Karpenter**: More efficient autoscaling than CA
4. **S3 Lifecycle**: Delete old work files
5. **Right-Sizing**: Use appropriate instance types

### Monitoring

1. **CloudWatch**: EKS cluster and S3 metrics
2. **Container Insights**: Pod-level metrics
3. **Prometheus**: Kubernetes metrics
4. **X-Ray**: Distributed tracing
5. **CloudTrail**: API audit logs

## Troubleshooting

### Common Issues

1. **Cannot connect to cluster**:
   - Verify AWS credentials have eks:DescribeCluster
   - Check cluster endpoint accessibility
   - Verify kubeconfig/credentials

2. **Pods cannot access S3**:
   - Verify IRSA role ARN annotation
   - Check IAM role trust policy
   - Verify S3 bucket permissions

3. **Pods fail to schedule**:
   - Check node selectors and tolerations
   - Verify node capacity
   - Check pod resource requests

4. **Fusion storage not working**:
   - Verify storage_mode = "fusion"
   - Check S3 bucket permissions
   - Verify Fusion v2 is installed in cluster

5. **Permission denied errors**:
   - Check service account RBAC
   - Verify SecurityContext UID/GID
   - Check S3 bucket policies

## References

- [Amazon EKS Documentation](https://docs.aws.amazon.com/eks/)
- [EKS Best Practices Guide](https://aws.github.io/aws-eks-best-practices/)
- [IRSA Documentation](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [Nextflow Kubernetes Executor](https://www.nextflow.io/docs/latest/kubernetes.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller/)
- [EBS CSI Driver](https://github.com/kubernetes-sigs/aws-ebs-csi-driver)
- [EFS CSI Driver](https://github.com/kubernetes-sigs/aws-efs-csi-driver)
- [Karpenter](https://karpenter.sh/)
