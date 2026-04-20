# Slurm Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_slurm` resource, which manages Slurm compute environments in Seqera Platform.

Slurm compute environments enable running Nextflow workflows on HPC (High Performance Computing) clusters managed by the Slurm Workload Manager, providing access to traditional supercomputing infrastructure.

## Key Characteristics

- **HPC-Focused**: Designed for traditional HPC clusters and supercomputers
- **On-Premises**: Works with on-premises and university HPC systems
- **Batch Scheduling**: Leverages Slurm's advanced scheduling and resource management
- **Shared Filesystem**: Requires POSIX-compliant shared filesystem (NFS, Lustre, GPFS, etc.)
- **Multi-User**: Support for multi-tenant HPC environments with fair-share scheduling
- **Advanced Resources**: Support for GPUs, MPI, specialized hardware

## Resource Structure

```hcl
resource "seqera_compute_slurm" "example" {
  name         = "slurm-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_slurm_credential.main.credentials_id

  # Work directory (must be on shared filesystem)
  work_directory = "/shared/nextflow/work"

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
    pre_run_script  = "#!/bin/bash\nmodule load nextflow"
    post_run_script = "#!/bin/bash\necho 'Workflow complete'"
  }

  # Nextflow configuration
  nextflow_config = <<-EOF
    process {
      executor = 'slurm'
      queue = 'compute'
    }
  EOF

  # Environment variables
  environment_variables = {
    "SLURM_CLUSTER" = "hpc-cluster"
  }

  # Advanced options
  advanced {
    max_queue_size              = 100
    head_queue                  = "short"
    compute_queue               = "compute"
    head_job_submit_options     = "--time=24:00:00 --mem=8G"
    propagate_head_job_submit_options = false
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
- **Example**: `"slurm-hpc"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: Slurm credentials ID to use for accessing the HPC cluster
- **Reference**: Must reference a valid `seqera_slurm_credential` resource
- **Example**: `seqera_slurm_credential.main.credentials_id`
- **Notes**: Credentials contain SSH connection details for the Slurm head node

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: Mount path for the Nextflow work directory
- **Format**: Absolute POSIX path on shared filesystem
- **Example**: `"/shared/nextflow/work"`, `"/gpfs/scratch/username/work"`, `"/lustre/project/workflows/work"`
- **Character Limit**: 0/200 characters
- **Constraints**:
  - Must be an absolute path (start with `/`)
  - Must be on a shared filesystem accessible from all compute nodes
  - User must have read/write permissions
  - Common filesystems: NFS, Lustre, GPFS, BeeGFS, Ceph

### Optional Fields

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
      value = "genomics-workspace"
    },
    {
      name  = "cluster"
      value = "hpc-prod"
    },
    {
      name  = "partition"
      value = "compute"
    }
  ]
  ```
- **Notes**:
  - Only one resource label with the same name can be used (API constraint)
  - Default resource labels are pre-filled
  - Used for tracking and organization

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
  # Load required modules
  module purge
  module load nextflow/23.10.0
  module load java/11
  module load singularity/3.8

  # Set up environment
  export NXF_SINGULARITY_CACHEDIR=/shared/singularity/cache
  export TMPDIR=/scratch/$USER/tmp

  # Validate shared filesystem
  if [ ! -d /shared/nextflow/work ]; then
    echo "ERROR: Work directory not accessible"
    exit 1
  fi

  echo "Environment initialized"
  ```
- **Use Cases**:
  - Load environment modules (common on HPC systems)
  - Set up temporary directories
  - Validate filesystem access
  - Configure Singularity/Apptainer
  - Set environment variables

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

  # Archive results
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  rsync -av /shared/nextflow/results/ /archive/results-$TIMESTAMP/

  # Cleanup scratch space
  rm -rf /scratch/$USER/tmp/*

  # Send notification
  echo "Workflow completed at $TIMESTAMP" | mail -s "Pipeline Complete" user@example.com
  ```
- **Use Cases**:
  - Archive results to long-term storage
  - Cleanup temporary files
  - Send email notifications
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
    executor = 'slurm'
    queue = 'compute'

    errorStrategy = 'retry'
    maxRetries = 3

    cpus = 4
    memory = '16 GB'
    time = '4h'

    withLabel: big_mem {
      memory = '128 GB'
      time = '24h'
      queue = 'highmem'
    }

    withLabel: gpu {
      queue = 'gpu'
      clusterOptions = '--gres=gpu:1'
    }

    withLabel: mpi {
      queue = 'mpi'
      clusterOptions = '--ntasks=16'
    }
  }

  executor {
    queueSize = 100
    pollInterval = '30 sec'
    queueStatInterval = '5 min'
  }

  singularity {
    enabled = true
    autoMounts = true
    cacheDir = '/shared/singularity/cache'
  }
  ```
- **Use Cases**:
  - Configure Slurm executor settings
  - Set default resource requirements
  - Define queue/partition mappings
  - Configure Singularity/Apptainer
  - Set retry strategies

### Environment Variables

#### `environment_variables`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables set in all workflow jobs
- **Example**:
  ```hcl
  environment_variables = {
    "SLURM_CLUSTER"             = "hpc-cluster"
    "NXF_ANSI_LOG"              = "false"
    "NXF_OPTS"                  = "-Xms1g -Xmx4g"
    "TMPDIR"                    = "/scratch/$USER/tmp"
    "NXF_SINGULARITY_CACHEDIR"  = "/shared/singularity/cache"
  }
  ```
- **Notes**:
  - Available to all processes in the workflow
  - Can reference environment variables with `$USER`, `$HOME`, etc.
  - Useful for configuring tools and runtime behavior

### Advanced Options Block

#### `advanced`
Advanced configuration options for Slurm integration.

##### `advanced.max_queue_size`
- **Type**: Integer
- **Optional**: Yes
- **Description**: Maximum number of jobs that can be queued/running at the same time
- **Default**: `100` (typical)
- **Range**: 1 to thousands (depends on cluster limits)
- **Example**: `500`
- **Character Limit**: Input field for number
- **Notes**:
  - Controls how many Slurm jobs Nextflow submits simultaneously
  - Should consider cluster fair-share policies
  - Higher values for large parallel workflows
  - Lower values to be a good citizen on shared clusters

##### `advanced.head_queue`
- **Type**: String
- **Optional**: Yes
- **Description**: Slurm partition/queue for the head job (Nextflow orchestrator)
- **Example**: `"short"`, `"interactive"`, `"serial"`, `"login"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Partition where the Nextflow head process runs
  - Should be a partition that allows long-running jobs
  - Consider using dedicated partition for orchestration
  - Different from compute_queue

##### `advanced.compute_queue`
- **Type**: String
- **Optional**: Yes
- **Description**: Default Slurm partition/queue for compute jobs
- **Example**: `"compute"`, `"batch"`, `"standard"`, `"normal"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Default partition for workflow tasks
  - Can be overridden per-process in Nextflow config
  - Should have sufficient resources for typical jobs
  - Consider fair-share policies

##### `advanced.head_job_submit_options`
- **Type**: String
- **Optional**: Yes
- **Description**: Slurm submit options for the head job (sbatch parameters)
- **Format**: Space-separated Slurm options without `sbatch` command
- **Examples**:
  - `"--time=24:00:00 --mem=8G"`
  - `"--time=48:00:00 --mem=16G --cpus-per-task=4"`
  - `"--time=72:00:00 --mem=32G --partition=long"`
- **Character Limit**: 0/200 characters
- **Notes**:
  - Options passed to `sbatch` for head job
  - Common options: `--time`, `--mem`, `--cpus-per-task`, `--partition`
  - Head job runs for entire workflow duration
  - Should request sufficient time for workflow completion

##### `advanced.propagate_head_job_submit_options`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Whether to propagate head job submit options to compute jobs
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - If true, options from `head_job_submit_options` are applied to all jobs
  - Usually false, as compute jobs need different resources
  - Use `clusterOptions` in Nextflow config for per-process options

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

1. **work_directory**: Must be absolute path starting with `/`
2. **max_queue_size**: Must be positive integer
3. **head_job_submit_options**: Must be valid Slurm options format
4. **resource_labels**: Each label must have both `name` and `value`

### Lifecycle Considerations

- **Create**: Configures Slurm compute environment in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields
- **Delete**: Removes compute environment (doesn't affect Slurm cluster)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `credentials_id`

### Mutable Fields

These fields can be updated without replacement:
- `work_directory`
- `resource_labels`
- `staging_options`
- `nextflow_config`
- `environment_variables`
- `advanced` options

### Sensitive Fields

- The referenced `credentials_id` points to sensitive SSH credentials
- Scripts may contain sensitive information
- Environment variables may contain secrets

## Examples

### Minimal Configuration

```hcl
resource "seqera_compute_slurm" "minimal" {
  name           = "slurm-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_slurm_credential.main.credentials_id
  work_directory = "/shared/nextflow/work"
}
```

### Standard Configuration

```hcl
resource "seqera_compute_slurm" "standard" {
  name           = "slurm-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_slurm_credential.main.credentials_id
  work_directory = "/gpfs/project/nextflow/work"

  resource_labels = [
    {
      name  = "cluster"
      value = "hpc-prod"
    },
    {
      name  = "partition"
      value = "compute"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      module load nextflow/23.10.0
      module load singularity/3.8
      export NXF_SINGULARITY_CACHEDIR=/shared/singularity/cache
    EOF
  }

  advanced {
    max_queue_size = 100
    compute_queue  = "compute"
  }
}
```

### Production with Multiple Partitions

```hcl
resource "seqera_compute_slurm" "production" {
  name           = "slurm-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_slurm_credential.main.credentials_id
  work_directory = "/lustre/scratch/nextflow/work"

  resource_labels = [
    {
      name  = "environment"
      value = "production"
    },
    {
      name  = "cluster"
      value = "supercomputer"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      # Load modules
      module purge
      module load nextflow/23.10.0
      module load java/11
      module load singularity/3.8
      module load parallel

      # Set up directories
      export TMPDIR=/scratch/$USER/tmp
      mkdir -p $TMPDIR
      export NXF_SINGULARITY_CACHEDIR=/shared/singularity/cache
      mkdir -p $NXF_SINGULARITY_CACHEDIR

      # Validate filesystem
      if [ ! -w /lustre/scratch/nextflow/work ]; then
        echo "ERROR: Cannot write to work directory"
        exit 1
      fi

      echo "Production environment initialized"
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      echo "Pipeline completed with status: $NXF_EXIT_STATUS"

      # Archive results
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      ARCHIVE_DIR=/archive/workflows/$TIMESTAMP
      mkdir -p $ARCHIVE_DIR
      rsync -av --progress /lustre/scratch/nextflow/results/ $ARCHIVE_DIR/

      # Cleanup scratch
      find /scratch/$USER/tmp -type f -mtime +7 -delete

      # Generate report
      cat > $ARCHIVE_DIR/summary.txt <<REPORT
      Workflow Summary
      ================
      Completion Time: $(date)
      Exit Status: $NXF_EXIT_STATUS
      Archive Location: $ARCHIVE_DIR
REPORT

      # Send notification
      echo "Workflow completed. Results archived to $ARCHIVE_DIR" | \
        mail -s "Pipeline Complete - $TIMESTAMP" user@example.com
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'slurm'
      queue = 'compute'

      errorStrategy = 'retry'
      maxRetries = 2

      // Default resources
      cpus = 4
      memory = '16 GB'
      time = '4h'

      // High memory processes
      withLabel: big_mem {
        memory = '128 GB'
        time = '24h'
        queue = 'highmem'
        clusterOptions = '--mem=128G'
      }

      // GPU processes
      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '--gres=gpu:v100:1'
        time = '8h'
      }

      // MPI processes
      withLabel: mpi {
        queue = 'mpi'
        clusterOptions = '--ntasks=32 --nodes=2'
        time = '12h'
      }

      // Fast/short jobs
      withLabel: short {
        queue = 'short'
        time = '30m'
      }
    }

    executor {
      queueSize = 200
      pollInterval = '30 sec'
      queueStatInterval = '5 min'
      submitRateLimit = '10 sec'
    }

    singularity {
      enabled = true
      autoMounts = true
      cacheDir = '/shared/singularity/cache'
      runOptions = '--bind /lustre:/lustre --bind /gpfs:/gpfs'
    }

    report {
      enabled = true
      file = 'pipeline_report.html'
    }

    trace {
      enabled = true
      file = 'pipeline_trace.txt'
    }

    timeline {
      enabled = true
      file = 'timeline.html'
    }
  EOF

  environment_variables = {
    "SLURM_CLUSTER"             = "supercomputer"
    "NXF_ANSI_LOG"              = "false"
    "NXF_OPTS"                  = "-Xms2g -Xmx8g"
    "TMPDIR"                    = "/scratch/$USER/tmp"
    "NXF_SINGULARITY_CACHEDIR"  = "/shared/singularity/cache"
  }

  advanced {
    max_queue_size              = 200
    head_queue                  = "serial"
    compute_queue               = "compute"
    head_job_submit_options     = "--time=48:00:00 --mem=16G --cpus-per-task=4"
    propagate_head_job_submit_options = false
  }
}
```

### GPU Workloads

```hcl
resource "seqera_compute_slurm" "gpu" {
  name           = "slurm-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_slurm_credential.main.credentials_id
  work_directory = "/gpfs/scratch/gpu-work"

  resource_labels = [
    {
      name  = "compute_type"
      value = "gpu"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      module load nextflow cuda/11.8 singularity
      export NXF_SINGULARITY_CACHEDIR=/shared/singularity/cache
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'slurm'
      queue = 'gpu'

      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '--gres=gpu:a100:2 --time=24:00:00'
        containerOptions = '--nv'
      }
    }

    singularity {
      enabled = true
      autoMounts = true
      runOptions = '--nv'
    }
  EOF

  environment_variables = {
    "CUDA_VISIBLE_DEVICES" = "0,1"
  }

  advanced {
    compute_queue = "gpu"
  }
}
```

### MPI/Parallel Workloads

```hcl
resource "seqera_compute_slurm" "mpi" {
  name           = "slurm-mpi"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_slurm_credential.main.credentials_id
  work_directory = "/lustre/project/mpi-work"

  resource_labels = [
    {
      name  = "workload_type"
      value = "mpi"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      module load nextflow openmpi/4.1 singularity
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'slurm'
      queue = 'mpi'

      withLabel: mpi {
        queue = 'mpi'
        clusterOptions = '--ntasks=64 --nodes=4 --ntasks-per-node=16'
        time = '12h'
      }
    }
  EOF

  advanced {
    compute_queue = "mpi"
  }
}
```

### High-Throughput Workloads

```hcl
resource "seqera_compute_slurm" "high_throughput" {
  name           = "slurm-highthroughput"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_slurm_credential.main.credentials_id
  work_directory = "/gpfs/scratch/high-throughput"

  resource_labels = [
    {
      name  = "workload_type"
      value = "high_throughput"
    }
  ]

  nextflow_config = <<-EOF
    process {
      executor = 'slurm'
      queue = 'short'

      errorStrategy = { task.exitStatus == 140 ? 'retry' : 'ignore' }
      maxRetries = 5

      // Small, fast jobs
      cpus = 1
      memory = '2 GB'
      time = '30m'
    }

    executor {
      queueSize = 1000
      pollInterval = '15 sec'
      submitRateLimit = '100/1min'
    }
  EOF

  advanced {
    max_queue_size = 1000
    compute_queue  = "short"
  }
}
```

### University HPC Cluster

```hcl
resource "seqera_compute_slurm" "university" {
  name           = "slurm-university"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_slurm_credential.main.credentials_id
  work_directory = "/home/$USER/nextflow/work"

  resource_labels = [
    {
      name  = "institution"
      value = "university"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      # Load standard modules
      module load nextflow java singularity

      # Respect fair-share policies
      export NXF_OPTS="-Xms512m -Xmx2g"
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'slurm'
      queue = 'standard'

      // Be conservative with resources
      cpus = 2
      memory = '4 GB'
      time = '2h'

      errorStrategy = 'retry'
      maxRetries = 2
    }

    executor {
      queueSize = 50
      pollInterval = '1 min'
    }
  EOF

  advanced {
    max_queue_size = 50
    compute_queue  = "standard"
    head_job_submit_options = "--time=24:00:00 --mem=4G"
  }
}
```

## Slurm-Specific Considerations

### Shared Filesystem Requirements

The work directory **must** be on a shared filesystem accessible from:
- The login/head node (where Nextflow runs)
- All compute nodes (where jobs execute)

Common shared filesystems:
- **NFS**: Simple, widely used
- **Lustre**: High-performance parallel filesystem
- **GPFS/Spectrum Scale**: IBM high-performance filesystem
- **BeeGFS**: Parallel cluster filesystem
- **CephFS**: Distributed filesystem
- **GlusterFS**: Scalable network filesystem

### Environment Modules

Most HPC systems use Environment Modules:
```bash
module purge
module load nextflow/23.10.0
module load java/11
module load singularity/3.8
```

Essential modules:
- **nextflow**: Workflow engine
- **java**: Required by Nextflow
- **singularity/apptainer**: Container runtime (Docker usually not available)
- **Other tools**: MPI, CUDA, compilers as needed

### Container Runtime

HPC clusters typically use Singularity/Apptainer instead of Docker:
- **Why**: Docker requires root access, security concerns
- **Singularity**: Designed for HPC, no privileged operations
- **Compatibility**: Can run Docker images
- **Configuration**:
  ```groovy
  singularity {
    enabled = true
    autoMounts = true
    cacheDir = '/shared/singularity/cache'
    runOptions = '--bind /lustre:/lustre'
  }
  ```

### Resource Specifications

Slurm uses specific terminology:
- **Partition/Queue**: Logical grouping of nodes
- **Time Limit**: `--time=HH:MM:SS` or `--time=D-HH:MM:SS`
- **Memory**: `--mem=16G` or `--mem-per-cpu=4G`
- **CPUs**: `--cpus-per-task=4`
- **GPUs**: `--gres=gpu:v100:2`
- **Nodes**: `--nodes=4`
- **Tasks**: `--ntasks=16` (for MPI)

### Slurm ClusterOptions

Per-process Slurm options in Nextflow config:
```groovy
process {
  withLabel: big_mem {
    clusterOptions = '--mem=128G --time=24:00:00'
  }

  withLabel: gpu {
    clusterOptions = '--gres=gpu:v100:1 --time=8:00:00'
  }

  withLabel: mpi {
    clusterOptions = '--ntasks=32 --nodes=2'
  }
}
```

### Fair-Share and QOS

Be a good citizen on shared clusters:
- Respect fair-share policies
- Use appropriate QOS (Quality of Service)
- Set reasonable `max_queue_size`
- Use `submitRateLimit` to avoid overwhelming scheduler
- Example:
  ```groovy
  executor {
    queueSize = 100
    submitRateLimit = '10 sec'
    pollInterval = '1 min'
  }
  ```

## Best Practices

### Resource Management

1. **Set Appropriate Limits**: Use reasonable defaults for CPU, memory, time
2. **Queue Selection**: Use appropriate partitions for different job types
3. **Max Queue Size**: Balance throughput with fair-share policies
4. **Retry Strategy**: Handle transient failures and pre-emption
5. **Time Limits**: Request sufficient time but not excessive

### Shared Filesystem

1. **Work Directory**: Use high-performance shared filesystem
2. **Scratch Space**: Utilize local scratch for temporary files
3. **Cleanup**: Regularly clean old work files
4. **Quotas**: Monitor filesystem quotas
5. **Permissions**: Ensure proper read/write permissions

### Container Usage

1. **Singularity**: Use Singularity/Apptainer instead of Docker
2. **Cache Directory**: Use shared cache to avoid re-pulling images
3. **Bind Mounts**: Mount necessary filesystems with `runOptions`
4. **Image Format**: Use SIF format for best performance
5. **Build Cache**: Pre-build images to avoid runtime delays

### Performance

1. **Parallel Execution**: Leverage Slurm's parallel capabilities
2. **Node Selection**: Use appropriate node types (CPU, GPU, high-mem)
3. **Local Scratch**: Use node-local storage for I/O-intensive tasks
4. **Network**: Ensure good connectivity to shared filesystem
5. **Job Size**: Balance job size with scheduler overhead

### Security

1. **Credentials**: Secure SSH credentials with proper permissions
2. **File Permissions**: Use appropriate umask and permissions
3. **Scratch Cleanup**: Clean sensitive data from scratch
4. **Audit Logs**: Monitor Slurm accounting for usage
5. **Access Control**: Use Slurm ACLs for multi-tenant environments

### Monitoring

1. **Slurm Accounting**: Use `sacct` to monitor job history
2. **Queue Status**: Check `squeue` for job status
3. **Resource Usage**: Monitor with `sstat` and `sacct`
4. **Nextflow Reports**: Enable timeline, trace, and report
5. **Filesystem**: Monitor I/O and space usage

## Troubleshooting

### Common Issues

1. **Jobs Not Starting**:
   - Check partition availability: `sinfo`
   - Verify fair-share limits: `sshare`
   - Check job priority: `sprio`
   - Review QOS limits

2. **Permission Denied**:
   - Verify work directory permissions
   - Check SSH credentials
   - Ensure filesystem is mounted on compute nodes

3. **Module Not Found**:
   - Verify module is available: `module avail`
   - Load required modules in pre-run script
   - Check module dependencies

4. **Singularity Errors**:
   - Verify Singularity is available on compute nodes
   - Check cache directory permissions
   - Ensure bind mounts are correct
   - Verify image format compatibility

5. **Out of Memory/Time**:
   - Increase memory allocation
   - Extend time limits
   - Use appropriate partition
   - Check job resource usage with `sacct`

6. **Shared Filesystem Issues**:
   - Verify filesystem is mounted
   - Check network connectivity
   - Monitor I/O performance
   - Check quotas: `quota -s`

## API Mapping

### Seqera Platform API Endpoints

- **Create**: `POST /compute-envs`
- **Read**: `GET /compute-envs/{computeEnvId}`
- **Update**: `PUT /compute-envs/{computeEnvId}`
- **Delete**: `DELETE /compute-envs/{computeEnvId}`
- **List**: `GET /compute-envs?workspaceId={workspaceId}`

### Request/Response Schema Mapping

- Resource `name` → API `config.name`
- Resource `work_directory` → API `config.workDir`
- Resource `resource_labels` → API `config.resourceLabels`
- Resource `staging_options.pre_run_script` → API `config.preRunScript`
- Resource `staging_options.post_run_script` → API `config.postRunScript`
- Resource `nextflow_config` → API `config.environment` or `config.nextflowConfig`
- Resource `advanced.max_queue_size` → API `config.maxQueueSize`
- Resource `advanced.head_queue` → API `config.headQueue`
- Resource `advanced.compute_queue` → API `config.computeQueue`
- Resource `advanced.head_job_submit_options` → API `config.headJobSubmitOptions`
- Resource `advanced.propagate_head_job_submit_options` → API `config.propagateHeadJobSubmitOptions`

### Platform Type

- **API Value**: `"slurm-platform"`
- **Config Type**: `"SlurmComputeConfig"`

## Related Resources

- `seqera_slurm_credential` - SSH credentials for Slurm head node
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Slurm Cluster Requirements

### Required Software

- **Slurm**: 20.x or later
- **Nextflow**: 22.x or later
- **Java**: 11 or later
- **Singularity/Apptainer**: 3.x or later (recommended)
- **Shared Filesystem**: NFS, Lustre, GPFS, etc.

### Network Requirements

- SSH access to Slurm head/login node
- Shared filesystem accessible from all nodes
- Outbound internet access (for pulling containers, if needed)

### User Requirements

- User account on HPC cluster
- SSH key-based authentication
- Read/write access to shared filesystem
- Slurm job submission permissions
- Appropriate fair-share allocation

## References

- [Slurm Workload Manager](https://slurm.schedmd.com/)
- [Nextflow Slurm Executor](https://www.nextflow.io/docs/latest/executor.html#slurm)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Singularity/Apptainer](https://apptainer.org/)
- [Environment Modules](https://modules.readthedocs.io/)
- [Slurm Accounting](https://slurm.schedmd.com/accounting.html)
- [Slurm SBATCH Options](https://slurm.schedmd.com/sbatch.html)
