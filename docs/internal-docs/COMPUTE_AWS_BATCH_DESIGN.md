# AWS Batch Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_aws_batch` resource, which manages AWS Batch compute environments in Seqera Platform.

AWS Batch compute environments provide scalable compute capacity for running Nextflow workflows on AWS using AWS Batch service.

## Resource Structure

```hcl
resource "seqera_compute_aws_batch" "example" {
  name         = "aws-batch-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"

  work_directory = "s3://my-bucket/work"

  # Compute queue configuration
  compute_queue {
    name = "my-batch-queue"
    cpus = 8
    memory = "32 GB"
  }

  # EFS configuration
  efs_filesystem {
    file_system_id = "fs-1234567890abcdef0"
    mount_path     = "/mnt/efs"
  }

  # FSx configuration
  fsx_filesystem {
    file_system_id = "fs-0123456789abcdef"
    mount_path     = "/fsx"
  }

  # Advanced options
  advanced {
    forge_type              = "SPOT"
    allocation_strategy     = "SPOT_CAPACITY_OPTIMIZED"
    ec2_key_pair            = "my-keypair"
    min_cpus                = 0
    max_cpus                = 256
    gpu_enabled             = false
    instance_types          = ["m5.xlarge", "m5.2xlarge"]
    ebs_auto_scale          = true
    ebs_block_size          = 100
    vpc_id                  = "vpc-1234567890abcdef0"
    subnets                 = ["subnet-12345", "subnet-67890"]
    security_groups         = ["sg-12345678"]
    use_public_ips          = false
    head_queue_cpus         = 4
    head_queue_memory       = "8 GB"
    compute_job_role        = "arn:aws:iam::123456789012:role/BatchJobRole"
    execution_role          = "arn:aws:iam::123456789012:role/BatchExecutionRole"
    cli_path                = "/home/ec2-user/miniconda/bin/aws"
    fusion_v2_enabled       = true
    wave_enabled            = true
    fargate_head_enabled    = false
    disposition             = "CREATE"
  }

  # Staging options
  staging_options {
    pre_run_script  = "#!/bin/bash\necho 'Starting workflow'"
    post_run_script = "#!/bin/bash\necho 'Workflow complete'"
    head_job_cpus   = 2
    head_job_memory = "4 GB"
  }

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
- **Constraints**: Must be unique within the workspace
- **Example**: `"aws-batch-prod"`

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

#### `region`

- **Type**: String
- **Required**: Yes
- **Description**: AWS region where the Batch compute environment will be created
- **Validation**: Must be a valid AWS region code
- **Examples**: `"us-east-1"`, `"eu-west-1"`, `"ap-southeast-2"`

#### `work_directory`

- **Type**: String
- **Required**: Yes
- **Description**: S3 bucket path for Nextflow work directory where intermediate files will be stored
- **Format**: `s3://bucket-name/path`
- **Example**: `"s3://my-nextflow-bucket/work"`
- **Notes**: Bucket must exist and credentials must have read/write access

### Optional Configuration Blocks

#### `compute_queue`

Container configuration for the AWS Batch compute queue.

##### `compute_queue.name`

- **Type**: String
- **Optional**: Yes
- **Description**: Name of the AWS Batch compute queue
- **Default**: Auto-generated if not specified
- **Example**: `"nextflow-batch-queue"`

##### `compute_queue.cpus`

- **Type**: Integer
- **Optional**: Yes
- **Description**: Number of CPUs allocated for the compute queue
- **Default**: Based on instance type
- **Example**: `8`

##### `compute_queue.memory`

- **Type**: String
- **Optional**: Yes
- **Description**: Memory allocation for the compute queue
- **Format**: Number with unit (GB, MB)
- **Example**: `"32 GB"`

#### `efs_filesystem`

Configuration for mounting Amazon EFS (Elastic File System).

##### `efs_filesystem.file_system_id`

- **Type**: String
- **Optional**: Yes
- **Description**: EFS file system ID to mount
- **Format**: `fs-` followed by hexadecimal characters
- **Example**: `"fs-1234567890abcdef0"`
- **Notes**: EFS must be in the same VPC and region

##### `efs_filesystem.mount_path`

- **Type**: String
- **Optional**: Yes
- **Description**: Path where EFS will be mounted in the container
- **Default**: `"/mnt/efs"`
- **Example**: `"/mnt/shared"`

#### `fsx_filesystem`

Configuration for mounting Amazon FSx for Lustre.

##### `fsx_filesystem.file_system_id`

- **Type**: String
- **Optional**: Yes
- **Description**: FSx for Lustre file system ID to mount
- **Format**: `fs-` followed by hexadecimal characters
- **Example**: `"fs-0123456789abcdef"`
- **Notes**: FSx must be in the same VPC and region

##### `fsx_filesystem.mount_path`

- **Type**: String
- **Optional**: Yes
- **Description**: Path where FSx will be mounted in the container
- **Default**: `"/fsx"`
- **Example**: `"/mnt/fsx"`

##### `fsx_filesystem.dns_name`

- **Type**: String
- **Optional**: Yes
- **Description**: DNS name for the FSx file system
- **Example**: `"fs-0123456789abcdef.fsx.us-east-1.amazonaws.com"`

##### `fsx_filesystem.mount_name`

- **Type**: String
- **Optional**: Yes
- **Description**: Mount name for FSx Lustre
- **Default**: Usually the file system ID without the `fs-` prefix
- **Example**: `"fsx"`

### Advanced Options Block

#### `advanced`

Advanced configuration options for fine-tuning the compute environment.

##### `advanced.forge_type`

- **Type**: String
- **Optional**: Yes
- **Description**: Type of compute instances to provision
- **Allowed Values**:
  - `"SPOT"` - Use EC2 Spot instances (cost-effective, can be interrupted)
  - `"EC2"` - Use On-Demand EC2 instances (reliable, higher cost)
  - `"FARGATE"` - Use AWS Fargate serverless compute
- **Default**: `"EC2"`
- **Example**: `"SPOT"`

##### `advanced.allocation_strategy`

- **Type**: String
- **Optional**: Yes
- **Description**: Strategy for allocating compute resources
- **Allowed Values**:
  - `"BEST_FIT"` - Selects instance type that best fits job requirements
  - `"BEST_FIT_PROGRESSIVE"` - Similar to BEST_FIT but widens search progressively
  - `"SPOT_CAPACITY_OPTIMIZED"` - For Spot instances, selects from pools with optimal capacity
- **Default**: Depends on forge_type
- **Example**: `"SPOT_CAPACITY_OPTIMIZED"`
- **Notes**: SPOT_CAPACITY_OPTIMIZED only valid when forge_type is SPOT

##### `advanced.ec2_key_pair`

- **Type**: String
- **Optional**: Yes
- **Description**: EC2 key pair name for SSH access to compute instances
- **Example**: `"my-keypair"`
- **Notes**: Key pair must exist in the specified region

##### `advanced.min_cpus`

- **Type**: Integer
- **Optional**: Yes
- **Description**: Minimum number of CPUs to maintain in the compute environment
- **Default**: `0`
- **Range**: 0 to max_cpus
- **Example**: `0`
- **Notes**: Setting to 0 allows environment to scale to zero when idle

##### `advanced.max_cpus`

- **Type**: Integer
- **Optional**: Yes
- **Description**: Maximum number of CPUs available in the compute environment
- **Default**: `256`
- **Range**: min_cpus to account limits
- **Example**: `512`
- **Notes**: Subject to AWS service quotas

##### `advanced.gpu_enabled`

- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable GPU support for compute instances
- **Default**: `false`
- **Example**: `true`
- **Notes**: When enabled, GPU-capable instance types will be selected

##### `advanced.instance_types`

- **Type**: List of Strings
- **Optional**: Yes
- **Description**: List of EC2 instance types to use
- **Default**: `["optimal"]` (AWS Batch selects appropriate instances)
- **Examples**:
  - `["m5.xlarge", "m5.2xlarge", "m5.4xlarge"]`
  - `["c5.2xlarge", "c5.4xlarge"]` (compute-optimized)
  - `["r5.xlarge", "r5.2xlarge"]` (memory-optimized)
  - `["p3.2xlarge"]` (GPU instances)
- **Notes**: Specifying multiple types allows Batch to select based on availability

##### `advanced.ebs_auto_scale`

- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable automatic EBS volume expansion
- **Default**: `false`
- **Example**: `true`
- **Notes**: When enabled, EBS volumes automatically expand as needed

##### `advanced.ebs_block_size`

- **Type**: Integer
- **Optional**: Yes
- **Description**: Size of EBS root volume in GB
- **Default**: `50`
- **Range**: 8 to 16384
- **Example**: `100`
- **Notes**: Minimum 8 GB, maximum 16 TB

##### `advanced.vpc_id`

- **Type**: String
- **Optional**: Yes
- **Description**: VPC ID where compute environment will be deployed
- **Format**: `vpc-` followed by hexadecimal characters
- **Example**: `"vpc-1234567890abcdef0"`
- **Notes**: Required if using existing VPC instead of default

##### `advanced.subnets`

- **Type**: List of Strings
- **Optional**: Yes
- **Description**: List of subnet IDs for compute instances
- **Format**: Each subnet ID starts with `subnet-`
- **Example**: `["subnet-12345", "subnet-67890", "subnet-abcde"]`
- **Notes**:
  - Subnets must be in the specified VPC
  - Use multiple subnets for high availability
  - Must have sufficient IP addresses

##### `advanced.security_groups`

- **Type**: List of Strings
- **Optional**: Yes
- **Description**: List of security group IDs to attach to compute instances
- **Format**: Each security group ID starts with `sg-`
- **Example**: `["sg-12345678", "sg-87654321"]`
- **Notes**: Security groups must allow necessary network access

##### `advanced.use_public_ips`

- **Type**: Boolean
- **Optional**: Yes
- **Description**: Assign public IP addresses to compute instances
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Set to `true` if instances need direct internet access
  - Not needed if using NAT gateway
  - Consider security implications

##### `advanced.head_queue_cpus`

- **Type**: Integer
- **Optional**: Yes
- **Description**: Number of CPUs allocated for the head job queue
- **Default**: `1`
- **Example**: `4`
- **Notes**: The head job orchestrates workflow execution

##### `advanced.head_queue_memory`

- **Type**: String
- **Optional**: Yes
- **Description**: Memory allocation for the head job queue
- **Format**: Number with unit (GB, MB)
- **Default**: `"1 GB"`
- **Example**: `"8 GB"`

##### `advanced.compute_job_role`

- **Type**: String
- **Optional**: Yes
- **Description**: IAM role ARN for compute jobs
- **Format**: `arn:aws:iam::account-id:role/role-name`
- **Example**: `"arn:aws:iam::123456789012:role/BatchJobRole"`
- **Notes**:
  - Jobs assume this role during execution
  - Must have permissions for S3, CloudWatch, etc.

##### `advanced.execution_role`

- **Type**: String
- **Optional**: Yes
- **Description**: IAM role ARN for Batch execution (pulling container images, writing logs)
- **Format**: `arn:aws:iam::account-id:role/role-name`
- **Example**: `"arn:aws:iam::123456789012:role/BatchExecutionRole"`
- **Notes**: Must have permissions for ECR, CloudWatch Logs

##### `advanced.cli_path`

- **Type**: String
- **Optional**: Yes
- **Description**: Path to AWS CLI on compute instances
- **Default**: `"/home/ec2-user/miniconda/bin/aws"`
- **Example**: `"/usr/local/bin/aws"`
- **Notes**: AWS CLI must be available at this path

##### `advanced.fusion_v2_enabled`

- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable Fusion v2 for virtual file system
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Fusion provides virtual file system for efficient S3 access
  - Improves performance by lazy loading files

##### `advanced.wave_enabled`

- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable Wave containers service
- **Default**: `false`
- **Example**: `true`
- **Notes**: Wave builds and manages container images on-demand

##### `advanced.fargate_head_enabled`

- **Type**: Boolean
- **Optional**: Yes
- **Description**: Use Fargate for head job instead of EC2
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - Reduces costs by running head job on serverless compute
  - Only applicable when using EC2 for worker jobs

##### `advanced.disposition`

- **Type**: String
- **Optional**: Yes
- **Description**: Disposition of AWS Batch compute environment
- **Allowed Values**:
  - `"CREATE"` - Create new Batch compute environment
  - `"CREATE_OR_UPDATE"` - Create if doesn't exist, update if exists
- **Default**: `"CREATE"`
- **Example**: `"CREATE_OR_UPDATE"`

##### `advanced.bidding_percentage`

- **Type**: Integer
- **Optional**: Yes (Required when forge_type is SPOT)
- **Description**: Maximum percentage of On-Demand price to pay for Spot instances
- **Range**: 1 to 100
- **Default**: `100` (pay up to full On-Demand price)
- **Example**: `60` (pay up to 60% of On-Demand price)
- **Notes**: Only applicable when forge_type is SPOT

### Staging Options Block

#### `staging_options`

Configuration for workflow staging and lifecycle scripts.

##### `staging_options.pre_run_script`

- **Type**: String
- **Optional**: Yes
- **Description**: Bash script to run before workflow execution begins
- **Format**: Multi-line bash script
- **Example**:
  ```bash
  #!/bin/bash
  echo "Setting up environment..."
  module load java/11
  ```
- **Use Cases**:
  - Load environment modules
  - Download reference data
  - Set up tools/dependencies

##### `staging_options.post_run_script`

- **Type**: String
- **Optional**: Yes
- **Description**: Bash script to run after workflow execution completes
- **Format**: Multi-line bash script
- **Example**:
  ```bash
  #!/bin/bash
  echo "Cleaning up..."
  rm -rf /tmp/work/*
  ```
- **Use Cases**:
  - Cleanup temporary files
  - Archive results
  - Send notifications

##### `staging_options.head_job_cpus`

- **Type**: Integer
- **Optional**: Yes
- **Description**: Number of CPUs for the head job
- **Default**: `1`
- **Example**: `2`
- **Notes**: Different from advanced.head_queue_cpus (queue vs job level)

##### `staging_options.head_job_memory`

- **Type**: String
- **Optional**: Yes
- **Description**: Memory allocation for the head job
- **Format**: Number with unit (GB, MB)
- **Default**: `"1 GB"`
- **Example**: `"4 GB"`

### Environment Variables

#### `environment_variables`

- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables to set in all compute jobs
- **Example**:
  ```hcl
  environment_variables = {
    "JAVA_OPTS"        = "-Xmx4g"
    "NXF_ANSI_LOG"     = "false"
    "MY_CUSTOM_VAR"    = "value"
  }
  ```
- **Notes**:
  - Variables are available to all processes in the workflow
  - Useful for configuring tools and runtime behavior

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
.
### `last_updated`

- **Type**: String (RFC3339 timestamp)
- **Computed**: Yes
- **Description**: Timestamp when the compute environment was last modified

## Implementation Notes

### Validation Rules

1. **region**: Must be a valid AWS region code
2. **work_directory**: Must start with `s3://`
3. **min_cpus**: Must be less than or equal to max_cpus
4. **max_cpus**: Must be greater than or equal to min_cpus
5. **bidding_percentage**: Required when forge_type is SPOT, must be 1-100
6. **allocation_strategy**: SPOT_CAPACITY_OPTIMIZED only valid with forge_type SPOT
7. **instance_types**: Each must be a valid EC2 instance type
8. **memory fields**: Must include unit (GB or MB)

### Lifecycle Considerations

- **Create**: Provisions AWS Batch compute environment and Seqera Platform configuration
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields (some fields may require replacement)
- **Delete**: Removes compute environment from Seqera Platform (AWS resources cleanup depends on disposition)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:

- `name`
- `region`
- `credentials_id`
- `advanced.vpc_id`
- `advanced.subnets`
- `advanced.disposition`

### Sensitive Fields

No fields are marked as sensitive in the compute environment resource itself, but:

- The referenced `credentials_id` points to sensitive AWS credentials
- Scripts may contain sensitive information and should be handled carefully

## Examples

### Minimal Configuration

```hcl
resource "seqera_compute_aws_batch" "minimal" {
  name           = "aws-batch-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://my-bucket/work"
}
```

### Spot Instances with Cost Optimization

```hcl
resource "seqera_compute_aws_batch" "spot" {
  name           = "aws-batch-spot"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://my-bucket/work"

  advanced {
    forge_type            = "SPOT"
    allocation_strategy   = "SPOT_CAPACITY_OPTIMIZED"
    bidding_percentage    = 70
    min_cpus              = 0
    max_cpus              = 512
    instance_types        = ["m5.xlarge", "m5.2xlarge", "c5.xlarge", "c5.2xlarge"]
    ebs_auto_scale        = true
  }
}
```

### Production with VPC and EFS

```hcl
resource "seqera_compute_aws_batch" "production" {
  name           = "aws-batch-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://prod-bucket/work"

  efs_filesystem {
    file_system_id = aws_efs_file_system.shared.id
    mount_path     = "/mnt/efs"
  }

  advanced {
    forge_type          = "EC2"
    min_cpus            = 8
    max_cpus            = 512
    vpc_id              = aws_vpc.main.id
    subnets             = aws_subnet.private[*].id
    security_groups     = [aws_security_group.batch.id]
    use_public_ips      = false
    compute_job_role    = aws_iam_role.batch_job.arn
    execution_role      = aws_iam_role.batch_execution.arn
    fusion_v2_enabled   = true
    wave_enabled        = true
  }

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      echo "Loading modules..."
      module load nextflow/23.10.0
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Archiving results..."
      aws s3 sync /tmp/results s3://archive-bucket/
    EOF
  }

  environment_variables = {
    "NXF_ANSI_LOG"  = "false"
    "NXF_OPTS"      = "-Xms1g -Xmx4g"
    "AWS_REGION"    = "us-east-1"
  }
}
```

### GPU-Enabled Compute

```hcl
resource "seqera_compute_aws_batch" "gpu" {
  name           = "aws-batch-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_aws_credential.main.credentials_id
  region         = "us-east-1"
  work_directory = "s3://gpu-bucket/work"

  advanced {
    forge_type      = "EC2"
    gpu_enabled     = true
    instance_types  = ["p3.2xlarge", "p3.8xlarge"]
    max_cpus        = 256
    ebs_block_size  = 200
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
- Resource `work_directory` -> API `config.workDir`
- Resource `advanced.forge_type` → API `config.forge.type`
- Resource `advanced.vpc_id` → API `config.forge.vpcId`
- etc.

## Related Resources

- `seqera_aws_credential` - AWS credentials used by the compute environment
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## References

- [AWS Batch Documentation](https://docs.aws.amazon.com/batch/)
- [Nextflow AWS Batch Executor](https://www.nextflow.io/docs/latest/awscloud.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
