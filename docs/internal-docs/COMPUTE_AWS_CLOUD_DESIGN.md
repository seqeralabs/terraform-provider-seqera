# AWS Cloud Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_aws_cloud` resource, which manages AWS Cloud (EC2-based) compute environments in Seqera Platform.

AWS Cloud compute environments provide compute capacity using direct EC2 instances managed by Seqera Platform, offering more control and flexibility compared to AWS Batch.

## Key Differences: AWS Cloud vs AWS Batch

| Feature             | AWS Cloud                                    | AWS Batch                   |
| ------------------- | -------------------------------------------- | --------------------------- |
| **Compute Service** | Direct EC2 instances                         | AWS Batch service           |
| **Management**      | Seqera manages instances                     | AWS manages batch jobs      |
| **Scaling**         | Manual or auto-scaling groups                | Automatic via AWS Batch     |
| **Control**         | More direct control over instances           | Higher-level abstraction    |
| **Use Case**        | Custom requirements, specific instance needs | Simplified batch processing |

## Resource Structure

```hcl
resource "seqera_compute_aws_cloud" "example" {
  name         = "aws-cloud-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"

  work_directory = "s3://my-bucket/work"

  # Allowed S3 buckets for data access
  allowed_buckets = [
    "s3://my-data-bucket",
    "s3://reference-data-bucket/*",
    "s3://results-bucket/project-name/*"
  ]

  # Advanced configuration
  advanced {
    instance_type   = "m5.xlarge"
    ami_id          = "ami-0c55b159cbfafe1f0"
    key_pair        = "my-keypair"
    vpc_id          = "vpc-1234567890abcdef0"
    subnets         = ["subnet-12345", "subnet-67890"]
    security_groups = ["sg-12345678"]

    instance_profile_arn = "arn:aws:iam::123456789012:instance-profile/SeqeraComputeProfile"
    boot_disk_size       = 100

    use_graviton = false
  }

  # Staging options
  staging_options {
    pre_run_script  = "#!/bin/bash\necho 'Starting workflow'"
    post_run_script = "#!/bin/bash\necho 'Workflow complete'"
  }

  # Nextflow configuration
  nextflow_config = <<-EOF
    process {
      executor = 'awsbatch'
      queue = 'default'
    }
  EOF

  # Environment variables
  environment_variables = {
    "MY_VAR" = "value"
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
- **Constraints**:
  - Use only alphanumeric characters, dashes, and underscores
  - Must be unique within the workspace
- **Example**: `"aws-cloud-prod"`

#### `workspace_id`

- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`

- **Type**: String
- **Required**: Yes
- **Description**: AWS credentials ID to use for accessing AWS services
- **Reference**: Must reference a valid `seqera_aws_credential` resource
- **Example**: `seqera_aws_credential.main.credentials_id`
- **Notes**: Credentials must have permissions to:
  - Launch and manage EC2 instances
  - Access S3 buckets
  - Manage security groups and networking

#### `region`

- **Type**: String
- **Required**: Yes
- **Description**: AWS region where EC2 instances will be launched
- **Validation**: Must be a valid AWS region code
- **Examples**: `"us-east-1"`, `"eu-west-1"`, `"ap-southeast-2"`
- **Notes**: All resources (VPC, subnets, AMI) must be in this region

#### `work_directory`

- **Type**: String
- **Required**: Yes
- **Description**: S3 bucket path for Nextflow work directory
- **Format**: `s3://bucket-name/path`
- **Example**: `"s3://my-nextflow-bucket/work"`
- **Constraints**:
  - Bucket must exist
  - Credentials must have read/write access
  - Should be in the same region for better performance

### Optional Fields

#### `allowed_buckets`

- **Type**: List of Strings
- **Optional**: Yes
- **Description**: List of S3 buckets or paths that the compute environment can access
- **Format**: Each entry can be:
  - `s3://bucket-name` - Entire bucket
  - `s3://bucket-name/*` - All objects in bucket (explicit)
  - `s3://bucket-name/path/*` - Specific path prefix
- **Default**: Work directory bucket only
- **Examples**:
  ```hcl
  allowed_buckets = [
    "s3://input-data-bucket",
    "s3://reference-genomes/*",
    "s3://results/project-a/*",
    "s3://shared-resources/databases/*"
  ]
  ```
- **Notes**:
  - Grants read-write permissions for the specified paths
  - Wildcards (`*`) indicate all objects under the path
  - Essential for data access control and security

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
  echo "Initializing environment..."
  module load tools/nextflow
  aws s3 sync s3://reference-data/ /mnt/reference/
  ```
- **Use Cases**:
  - Load environment modules
  - Download reference data
  - Configure tools
  - Set up directories
  - Validate prerequisites

##### `staging_options.post_run_script`

- **Type**: String
- **Optional**: Yes
- **Description**: Bash script executed after workflow completes
- **Format**: Multi-line bash script
- **Character Limit**: 0/1024 characters
- **Example**:
  ```bash
  #!/bin/bash
  echo "Cleaning up..."
  rm -rf /tmp/work/*
  aws s3 sync /results/ s3://archive-bucket/$(date +%Y-%m-%d)/
  ```
- **Use Cases**:
  - Cleanup temporary files
  - Archive results
  - Send notifications
  - Generate reports
  - Update databases

##### `staging_options.nextflow_config`

- **Type**: String
- **Optional**: Yes
- **Description**: Custom Nextflow configuration for this compute environment
- **Format**: Nextflow configuration DSL
- **Character Limit**: 0/3200 characters
- **Example**:

  ```groovy
  process {
    executor = 'awsbatch'
    queue = 'default'

    errorStrategy = 'retry'
    maxRetries = 3

    cpus = 2
    memory = '4 GB'

    withLabel: big_mem {
      memory = '32 GB'
    }
  }

  aws {
    region = 'us-east-1'
    batch {
      cliPath = '/home/ec2-user/miniconda/bin/aws'
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
    "JAVA_OPTS"           = "-Xmx4g"
    "NXF_ANSI_LOG"        = "false"
    "NXF_OPTS"            = "-Xms1g -Xmx4g"
    "AWS_DEFAULT_REGION"  = "us-east-1"
    "CUSTOM_TOOL_PATH"    = "/opt/tools/bin"
  }
  ```
- **Notes**:
  - Available to all processes in the workflow
  - Useful for configuring tools and runtime behavior
  - Can override default Nextflow settings

### Advanced Options Block

#### `advanced`

Advanced configuration for EC2 instances and networking.

##### `advanced.instance_type`

- **Type**: String
- **Optional**: Yes
- **Description**: EC2 instance type to use for compute
- **Default**: Platform-determined based on workload
- **Examples**:
  - General Purpose: `"m5.xlarge"`, `"m5.2xlarge"`, `"m5.4xlarge"`
  - Compute Optimized: `"c5.2xlarge"`, `"c5.4xlarge"`, `"c5.9xlarge"`
  - Memory Optimized: `"r5.xlarge"`, `"r5.2xlarge"`, `"r5.4xlarge"`
  - GPU Instances: `"p3.2xlarge"`, `"p3.8xlarge"`, `"g4dn.xlarge"`
  - Burstable: `"t3.medium"`, `"t3.large"`, `"t3.xlarge"`
- **Notes**:
  - Must be available in the specified region
  - Consider CPU, memory, and network requirements
  - GPU instances for deep learning workloads

##### `advanced.use_graviton`

- **Type**: Boolean
- **Optional**: Yes
- **Description**: Use AWS Graviton (ARM64) CPU architecture
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Graviton instances offer better price/performance
  - Requires ARM64-compatible container images
  - Instance types: `m6g`, `c6g`, `r6g` series
  - Not compatible with all software

##### `advanced.ami_id`

- **Type**: String
- **Optional**: Yes
- **Description**: Custom Amazon Machine Image (AMI) ID
- **Format**: `ami-` followed by hexadecimal characters
- **Example**: `"ami-0c55b159cbfafe1f0"`
- **Default**: Seqera-provided AMI
- **Notes**:
  - AMI must be in the specified region
  - Must have required dependencies (Docker, AWS CLI, etc.)
  - Use for custom tools or configurations

##### `advanced.key_pair`

- **Type**: String
- **Optional**: Yes
- **Description**: EC2 key pair name for SSH access
- **Example**: `"my-keypair"`
- **Notes**:
  - Key pair must exist in the specified region
  - Used for debugging and troubleshooting
  - SSH access requires appropriate security group rules

##### `advanced.vpc_id`

- **Type**: String
- **Optional**: Yes
- **Description**: VPC ID where instances will be launched
- **Format**: `vpc-` followed by hexadecimal characters
- **Example**: `"vpc-1234567890abcdef0"`
- **Notes**:
  - VPC must be in the specified region
  - Required if specifying subnets
  - Default VPC used if not specified

##### `advanced.subnets`

- **Type**: List of Strings
- **Optional**: Yes
- **Description**: List of subnet IDs for launching instances
- **Format**: Each subnet ID starts with `subnet-`
- **Example**: `["subnet-12345678", "subnet-87654321", "subnet-abcdef01"]`
- **Notes**:
  - Subnets must be in the specified VPC
  - Use multiple subnets for high availability
  - Must have sufficient IP addresses
  - Consider public vs private subnets for internet access

##### `advanced.security_groups`

- **Type**: List of Strings
- **Optional**: Yes
- **Description**: List of security group IDs to attach to instances
- **Format**: Each security group ID starts with `sg-`
- **Example**: `["sg-12345678", "sg-87654321"]`
- **Notes**:
  - Security groups must be in the specified VPC
  - Must allow necessary outbound traffic (S3, Docker Hub, etc.)
  - Consider allowing SSH (port 22) for debugging

##### `advanced.instance_profile_arn`

- **Type**: String
- **Optional**: Yes
- **Description**: IAM instance profile ARN for EC2 instances
- **Format**: `arn:aws:iam::account-id:instance-profile/profile-name`
- **Example**: `"arn:aws:iam::123456789012:instance-profile/SeqeraComputeProfile"`
- **Notes**:
  - Instance profile must have permissions for:
    - S3 access (read/write to work directory and allowed buckets)
    - ECR access (pull container images)
    - CloudWatch Logs (write logs)
  - Instances assume this role during execution

##### `advanced.boot_disk_size`

- **Type**: Integer
- **Optional**: Yes
- **Description**: Size of root EBS volume in GB
- **Default**: `50`
- **Range**: 8 to 16384 (16 TB)
- **Example**: `100`
- **Notes**:
  - Minimum 8 GB required
  - Consider space for:
    - Container images
    - Temporary files
    - Logs
  - Larger workflows may need more space

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
- **Possible Values**:
  - `"CREATING"` - Being created
  - `"AVAILABLE"` - Ready for use
  - `"INVALID"` - Configuration error
  - `"DELETING"` - Being deleted

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

1. **region**: Must be a valid AWS region code
2. **work_directory**: Must start with `s3://`
3. **allowed_buckets**: Each entry must start with `s3://`
4. **instance_type**: Must be a valid EC2 instance type
5. **ami_id**: Must be valid AMI format (`ami-[a-f0-9]+`)
6. **vpc_id**: Must be valid VPC format (`vpc-[a-f0-9]+`)
7. **subnets**: Each must be valid subnet format (`subnet-[a-f0-9]+`)
8. **security_groups**: Each must be valid SG format (`sg-[a-f0-9]+`)
9. **boot_disk_size**: Must be between 8 and 16384

### Required IAM Permissions

#### For Credentials (AWS User/Role)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus",
        "ec2:DescribeInstanceTypes",
        "ec2:DescribeImages",
        "ec2:DescribeKeyPairs",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcs",
        "ec2:CreateTags",
        "iam:PassRole"
      ],
      "Resource": "*"
    }
  ]
}
```

#### For Instance Profile

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:ListBucket",
        "s3:DeleteObject"
      ],
      "Resource": [
        "arn:aws:s3:::work-bucket/*",
        "arn:aws:s3:::allowed-bucket-1/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
```

### Lifecycle Considerations

- **Create**: Provisions EC2 instance configuration in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields (some may require replacement)
- **Delete**: Removes compute environment (terminates any running instances)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:

- `name`
- `region`
- `credentials_id`
- `advanced.vpc_id`
- `advanced.subnets`

### Mutable Fields

These fields can be updated without replacement:

- `work_directory`
- `allowed_buckets`
- `staging_options`
- `nextflow_config`
- `environment_variables`
- `advanced.instance_type`
- `advanced.ami_id`
- `advanced.key_pair`
- `advanced.security_groups`
- `advanced.instance_profile_arn`
- `advanced.boot_disk_size`

## Examples

### Minimal Configuration

```hcl
resource "seqera_compute_aws_cloud" "minimal" {
  name           = "aws-cloud-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://my-bucket/work"
}
```

### Standard Configuration with Data Access

```hcl
resource "seqera_compute_aws_cloud" "standard" {
  name           = "aws-cloud-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://work-bucket/workflows"

  # Grant access to data buckets
  allowed_buckets = [
    "s3://input-data",
    "s3://reference-genomes/*",
    "s3://results/project-a/*"
  ]

  advanced {
    instance_type = "m5.2xlarge"
    key_pair      = "my-keypair"
  }
}
```

### Production Configuration with VPC and Custom AMI

```hcl
resource "seqera_compute_aws_cloud" "production" {
  name           = "aws-cloud-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://prod-work-bucket/workflows"

  allowed_buckets = [
    "s3://prod-input-data/*",
    "s3://prod-reference-data/*",
    "s3://prod-results/*",
    "s3://shared-resources/databases/*"
  ]

  advanced {
    instance_type        = "c5.4xlarge"
    ami_id               = "ami-0c55b159cbfafe1f0"
    key_pair             = "prod-keypair"
    vpc_id               = aws_vpc.main.id
    subnets              = aws_subnet.private[*].id
    security_groups      = [aws_security_group.compute.id]
    instance_profile_arn = aws_iam_instance_profile.compute.arn
    boot_disk_size       = 200
  }

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Loading environment for production workflow..."
      module load nextflow/23.10.0
      module load aws-cli/2.13.0

      # Pre-download reference data
      aws s3 sync s3://prod-reference-data/genome /mnt/reference/
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Archiving workflow results..."
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      aws s3 sync /results/ s3://archive-bucket/$TIMESTAMP/

      # Send notification
      aws sns publish --topic-arn arn:aws:sns:us-east-1:123456789012:workflow-complete \
        --message "Workflow completed: $TIMESTAMP"
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'awsbatch'
      errorStrategy = 'retry'
      maxRetries = 2

      cpus = 4
      memory = '16 GB'

      withLabel: intensive {
        cpus = 16
        memory = '64 GB'
      }
    }

    aws {
      region = 'us-east-1'
      batch.cliPath = '/usr/local/bin/aws'
    }
  EOF

  environment_variables = {
    "NXF_ANSI_LOG"       = "false"
    "NXF_OPTS"           = "-Xms2g -Xmx8g"
    "AWS_DEFAULT_REGION" = "us-east-1"
    "TMPDIR"             = "/scratch"
  }
}
```

### Graviton (ARM64) Configuration

```hcl
resource "seqera_compute_aws_cloud" "graviton" {
  name           = "aws-cloud-graviton"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://work-bucket/graviton"

  advanced {
    instance_type = "m6g.4xlarge"
    use_graviton  = true
    boot_disk_size = 150
  }

  environment_variables = {
    "NXF_OPTS" = "-Xms2g -Xmx8g"
    "ARCH"     = "arm64"
  }
}
```

### GPU-Enabled Configuration

```hcl
resource "seqera_compute_aws_cloud" "gpu" {
  name           = "aws-cloud-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://gpu-work-bucket/workflows"

  allowed_buckets = [
    "s3://ml-training-data/*",
    "s3://ml-models/*"
  ]

  advanced {
    instance_type  = "p3.2xlarge"
    ami_id         = "ami-0a1b2c3d4e5f6g7h8"  # GPU-optimized AMI
    boot_disk_size = 250
  }

  nextflow_config = <<-EOF
    process {
      withLabel: gpu {
        containerOptions = '--gpus all'
        clusterOptions = '--gres=gpu:1'
      }
    }
  EOF

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0"
    "NVIDIA_VISIBLE_DEVICES" = "all"
  }
}
```

### Multi-AZ High Availability Configuration

```hcl
resource "seqera_compute_aws_cloud" "ha" {
  name           = "aws-cloud-ha"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://ha-work-bucket/workflows"

  advanced {
    instance_type = "m5.xlarge"
    vpc_id        = aws_vpc.main.id

    # Spread across multiple availability zones
    subnets = [
      aws_subnet.private_us_east_1a.id,
      aws_subnet.private_us_east_1b.id,
      aws_subnet.private_us_east_1c.id
    ]

    security_groups      = [aws_security_group.compute.id]
    instance_profile_arn = aws_iam_instance_profile.compute.arn
  }
}
```

## Integration with Terraform AWS Provider

### Complete Example with Infrastructure

```hcl
# VPC Configuration
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "seqera-compute-vpc"
  }
}

# Subnets
resource "aws_subnet" "private" {
  count             = 3
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "seqera-private-${count.index + 1}"
  }
}

# Security Group
resource "aws_security_group" "compute" {
  name        = "seqera-compute"
  description = "Security group for Seqera compute instances"
  vpc_id      = aws_vpc.main.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow SSH from bastion
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
}

# IAM Role for Instances
resource "aws_iam_role" "compute" {
  name = "seqera-compute-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

# IAM Policy for S3 Access
resource "aws_iam_role_policy" "s3_access" {
  name = "s3-access"
  role = aws_iam_role.compute.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket",
          "s3:DeleteObject"
        ]
        Resource = [
          "arn:aws:s3:::work-bucket/*",
          "arn:aws:s3:::work-bucket"
        ]
      }
    ]
  })
}

# Instance Profile
resource "aws_iam_instance_profile" "compute" {
  name = "seqera-compute-profile"
  role = aws_iam_role.compute.name
}

# Seqera Compute Environment
resource "seqera_compute_aws_cloud" "integrated" {
  name           = "aws-cloud-integrated"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://work-bucket/workflows"

  advanced {
    instance_type        = "m5.2xlarge"
    vpc_id               = aws_vpc.main.id
    subnets              = aws_subnet.private[*].id
    security_groups      = [aws_security_group.compute.id]
    instance_profile_arn = aws_iam_instance_profile.compute.arn
    boot_disk_size       = 100
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

### Platform Type

- **API Value**: `"aws-batch"` (reuses AWS Batch platform with EC2 configuration)
- **Discriminator**: Configuration determines EC2 vs Batch behavior

## Related Resources

- `seqera_aws_credential` - AWS credentials used by the compute environment
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Comparison with AWS Batch

| Aspect               | AWS Cloud (EC2)                            | AWS Batch                             |
| -------------------- | ------------------------------------------ | ------------------------------------- |
| **Setup Complexity** | Simpler                                    | More complex                          |
| **Cost Control**     | Direct instance costs                      | Batch overhead                        |
| **Customization**    | Full control over instances                | Limited to Batch capabilities         |
| **Scaling**          | Manual or ASG                              | Automatic                             |
| **Management**       | Seqera manages                             | AWS manages                           |
| **Best For**         | Custom requirements, predictable workloads | Variable workloads, automatic scaling |

## References

- [AWS EC2 Documentation](https://docs.aws.amazon.com/ec2/)
- [Nextflow AWS Documentation](https://www.nextflow.io/docs/latest/awscloud.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
