# Moab Workload Manager Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_moab` resource, which manages Moab Workload Manager compute environments in Seqera Platform.

Moab compute environments enable running Nextflow workflows on HPC (High Performance Computing) clusters managed by Moab Workload Manager, providing advanced scheduling, resource management, and job orchestration for traditional supercomputing infrastructure.

## Key Characteristics

- **HPC-Focused**: Designed for traditional HPC clusters and supercomputers
- **On-Premises**: Works with on-premises and university HPC systems
- **Advanced Scheduling**: Leverages Moab's sophisticated scheduling algorithms
- **Shared Filesystem**: Requires POSIX-compliant shared filesystem (NFS, Lustre, GPFS, etc.)
- **Policy-Based**: Support for complex policies, priorities, and fairness
- **Integration**: Works with various resource managers (PBS, Torque, Slurm)

## Prerequisites

To connect Seqera Cloud to your local cluster and launch pipeline executions, the following requirements must be fulfilled:

1. **SSH Access**: The cluster must be reachable via an SSH connection using an SSH key
2. **Outbound Connectivity**: The cluster must allow outbound connections to the Seqera Cloud web service
3. **Job Submission**: The cluster queue used to run the Nextflow head job must be able to submit cluster jobs
4. **Nextflow Version**: Nextflow runtime version **22.10.0** (or later) must be pre-installed in the cluster

## Resource Structure

```hcl
resource "seqera_compute_moab" "example" {
  name         = "moab-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_moab_credential.main.credentials_id

  # Work directory (must be on shared filesystem)
  work_directory = "/shared/nextflow/work"

  # Optional: Launch directory
  launch_directory = "/shared/nextflow/launch"

  # Queue configuration
  head_queue_name    = "interactive"
  compute_queue_name = "compute"

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
      executor = 'moab'
      queue = 'compute'
    }
  EOF

  # Environment variables
  environment_variables = {
    "MOAB_CLUSTER" = "hpc-cluster"
  }

  # Advanced options
  advanced {
    nextflow_queue_size         = 100
    head_job_submit_options     = "-l walltime=24:00:00,mem=8gb"
    apply_head_job_submit_options = false
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
- **Example**: `"moab-hpc"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: Moab credentials ID to use for accessing the HPC cluster
- **Reference**: Must reference a valid `seqera_moab_credential` resource
- **Example**: `seqera_moab_credential.main.credentials_id`
- **Options**: Can be SSH-based or Tower agent-based credentials
- **Notes**: Credentials contain SSH connection details for the Moab cluster

#### `work_directory`
- **Type**: String
- **Required**: Yes
- **Description**: The Nextflow work directory on the cluster's shared file system
- **Format**: Absolute POSIX path on shared filesystem
- **Example**: `"/shared/nextflow/work"`, `"/gpfs/scratch/username/work"`, `"/lustre/project/workflows/work"`
- **Character Limit**: 0/200 characters
- **Constraints**:
  - Must be an absolute path (start with `/`)
  - Must be on a shared filesystem accessible from all compute nodes
  - User must have read-write access
  - Common filesystems: NFS, Lustre, GPFS, BeeGFS, Ceph

### Optional Fields

#### `launch_directory`
- **Type**: String
- **Optional**: Yes
- **Description**: The directory where Nextflow runs
- **Format**: Absolute POSIX path
- **Example**: `"/shared/nextflow/launch"`, `"/home/username/nextflow"`
- **Character Limit**: 0/200 characters
- **Default**: If omitted, defaults to the work directory
- **Constraints**:
  - Must be an absolute path
  - User must have read-write access
- **Notes**: Separate from work directory for organizing execution vs. intermediate files

#### `head_queue_name`
- **Type**: String
- **Optional**: Yes
- **Description**: The name of the queue on the cluster used to launch the execution of the Nextflow pipeline
- **Example**: `"interactive"`, `"serial"`, `"submit"`, `"launch"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Queue for the Nextflow head/orchestrator job
  - Should support long-running jobs
  - Must have job submission capabilities

#### `compute_queue_name`
- **Type**: String
- **Optional**: Yes
- **Description**: The name of queue on the cluster to which pipeline jobs are submitted
- **Example**: `"compute"`, `"batch"`, `"standard"`, `"normal"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Default queue for workflow tasks
  - Can be overridden by the pipeline configuration
  - Should have sufficient resources for typical jobs

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
      value = "moab-prod"
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
  mkdir -p $TMPDIR

  # Validate shared filesystem
  if [ ! -d /shared/nextflow/work ]; then
    echo "ERROR: Work directory not accessible"
    exit 1
  fi

  echo "Environment initialized"
  ```
- **Use Cases**:
  - Load environment modules
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
  echo "Workflow completed at $TIMESTAMP with status $NXF_EXIT_STATUS" | \
    mail -s "Pipeline Complete" user@example.com
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
    executor = 'moab'
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
      clusterOptions = '-l nodes=1:ppn=4:gpus=1'
    }

    withLabel: mpi {
      queue = 'mpi'
      clusterOptions = '-l nodes=4:ppn=16'
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
  - Configure Moab executor settings
  - Set default resource requirements
  - Define queue/class mappings
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
    "MOAB_CLUSTER"              = "hpc-cluster"
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
Advanced configuration options for Moab integration.

##### `advanced.nextflow_queue_size`
- **Type**: Integer
- **Optional**: Yes
- **Description**: The maximum number of jobs Nextflow can submit to the queue simultaneously
- **Default**: `100`
- **Range**: 1 to thousands (depends on cluster limits)
- **Example**: `500`
- **Notes**:
  - Controls how many Moab jobs Nextflow submits simultaneously
  - Should consider cluster fair-share policies
  - Higher values for large parallel workflows
  - Lower values to be a good citizen on shared clusters

##### `advanced.head_job_submit_options`
- **Type**: String
- **Optional**: Yes
- **Description**: Grid Engine submit options for the Nextflow head job
- **Format**: Moab/PBS options (qsub parameters) without the `qsub` command
- **Examples**:
  - `"-l walltime=24:00:00,mem=8gb"`
  - `"-l walltime=48:00:00,mem=16gb,nodes=1:ppn=4"`
  - `"-l walltime=72:00:00 -q long -A project123"`
- **Character Limit**: 0/200 characters
- **Notes**:
  - Options are added to the 'qsub' command run by Seqera Cloud to launch the pipeline execution
  - Common options: `-l walltime`, `-l mem`, `-l nodes`, `-q`, `-A` (account)
  - Head job runs for entire workflow duration
  - Should request sufficient time for workflow completion

##### `advanced.apply_head_job_submit_options`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable this field so the "Head job submit options" can also be applied to the Nextflow-submitted compute jobs
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
2. **launch_directory**: Must be absolute path starting with `/` if specified
3. **nextflow_queue_size**: Must be positive integer
4. **head_job_submit_options**: Must be valid Moab/PBS options format
5. **resource_labels**: Each label must have both `name` and `value`

### Lifecycle Considerations

- **Create**: Configures Moab compute environment in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields
- **Delete**: Removes compute environment (doesn't affect Moab cluster)

### Force Replacement Fields

The following fields require replacing the compute environment if changed:
- `name`
- `credentials_id`

### Mutable Fields

These fields can be updated without replacement:
- `work_directory`
- `launch_directory`
- `head_queue_name`
- `compute_queue_name`
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
resource "seqera_compute_moab" "minimal" {
  name           = "moab-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_moab_credential.main.credentials_id
  work_directory = "/shared/nextflow/work"
}
```

### Standard Configuration

```hcl
resource "seqera_compute_moab" "standard" {
  name           = "moab-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_moab_credential.main.credentials_id
  work_directory = "/gpfs/project/nextflow/work"
  launch_directory = "/gpfs/project/nextflow/launch"

  head_queue_name    = "interactive"
  compute_queue_name = "compute"

  resource_labels = [
    {
      name  = "cluster"
      value = "moab-hpc"
    },
    {
      name  = "environment"
      value = "production"
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
    nextflow_queue_size = 100
    head_job_submit_options = "-l walltime=24:00:00,mem=8gb"
  }
}
```

### Production with Multiple Queues

```hcl
resource "seqera_compute_moab" "production" {
  name           = "moab-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_moab_credential.main.credentials_id
  work_directory = "/lustre/scratch/nextflow/work"
  launch_directory = "/lustre/scratch/nextflow/launch"

  head_queue_name    = "serial"
  compute_queue_name = "batch"

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
      executor = 'moab'
      queue = 'batch'

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
        clusterOptions = '-l mem=128gb,walltime=24:00:00'
      }

      // GPU processes
      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '-l nodes=1:ppn=4:gpus=1,walltime=8:00:00'
      }

      // MPI processes
      withLabel: mpi {
        queue = 'mpi'
        clusterOptions = '-l nodes=4:ppn=16,walltime=12:00:00'
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
    "MOAB_CLUSTER"              = "supercomputer"
    "NXF_ANSI_LOG"              = "false"
    "NXF_OPTS"                  = "-Xms2g -Xmx8g"
    "TMPDIR"                    = "/scratch/$USER/tmp"
    "NXF_SINGULARITY_CACHEDIR"  = "/shared/singularity/cache"
  }

  advanced {
    nextflow_queue_size         = 200
    head_job_submit_options     = "-l walltime=48:00:00,mem=16gb,nodes=1:ppn=4 -q serial"
    apply_head_job_submit_options = false
  }
}
```

### GPU Workloads

```hcl
resource "seqera_compute_moab" "gpu" {
  name           = "moab-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_moab_credential.main.credentials_id
  work_directory = "/gpfs/scratch/gpu-work"

  head_queue_name    = "serial"
  compute_queue_name = "gpu"

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
      executor = 'moab'
      queue = 'gpu'

      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '-l nodes=1:ppn=4:gpus=2,walltime=24:00:00'
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
}
```

### MPI/Parallel Workloads

```hcl
resource "seqera_compute_moab" "mpi" {
  name           = "moab-mpi"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_moab_credential.main.credentials_id
  work_directory = "/lustre/project/mpi-work"

  head_queue_name    = "serial"
  compute_queue_name = "mpi"

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
      executor = 'moab'
      queue = 'mpi'

      withLabel: mpi {
        queue = 'mpi'
        clusterOptions = '-l nodes=8:ppn=16,walltime=12:00:00'
      }
    }
  EOF
}
```

### High-Throughput Workloads

```hcl
resource "seqera_compute_moab" "high_throughput" {
  name           = "moab-highthroughput"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_moab_credential.main.credentials_id
  work_directory = "/gpfs/scratch/high-throughput"

  head_queue_name    = "serial"
  compute_queue_name = "short"

  resource_labels = [
    {
      name  = "workload_type"
      value = "high_throughput"
    }
  ]

  nextflow_config = <<-EOF
    process {
      executor = 'moab'
      queue = 'short'

      errorStrategy = { task.exitStatus == 271 ? 'retry' : 'ignore' }
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
    nextflow_queue_size = 1000
  }
}
```

### University HPC Cluster

```hcl
resource "seqera_compute_moab" "university" {
  name           = "moab-university"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_moab_credential.main.credentials_id
  work_directory = "/home/$USER/nextflow/work"
  launch_directory = "/home/$USER/nextflow/launch"

  head_queue_name    = "interactive"
  compute_queue_name = "standard"

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
      executor = 'moab'
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
    nextflow_queue_size     = 50
    head_job_submit_options = "-l walltime=24:00:00,mem=4gb -A research_account"
  }
}
```

## Moab-Specific Considerations

### Shared Filesystem Requirements

The work directory **must** be on a shared filesystem accessible from:
- The login/submit node (where Nextflow runs)
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
- **nextflow**: Workflow engine (version 22.10.0 or later required)
- **java**: Required by Nextflow
- **singularity/apptainer**: Container runtime
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

### Moab/PBS Resource Specifications

Moab typically uses PBS/Torque syntax:
- **Walltime**: `-l walltime=HH:MM:SS` or `-l walltime=D:HH:MM:SS`
- **Memory**: `-l mem=16gb` or `-l pmem=4gb` (per process)
- **Nodes/Cores**: `-l nodes=4:ppn=16` (4 nodes, 16 cores per node)
- **GPUs**: `-l nodes=1:ppn=4:gpus=2`
- **Queue**: `-q queuename`
- **Account**: `-A projectname` or `-W group_list=projectname`

### Moab ClusterOptions

Per-process Moab/PBS options in Nextflow config:
```groovy
process {
  withLabel: big_mem {
    clusterOptions = '-l mem=128gb,walltime=24:00:00 -q highmem'
  }

  withLabel: gpu {
    clusterOptions = '-l nodes=1:ppn=4:gpus=1,walltime=8:00:00 -q gpu'
  }

  withLabel: mpi {
    clusterOptions = '-l nodes=4:ppn=16,walltime=12:00:00 -q mpi'
  }
}
```

### Job Classes and Queues

Moab supports sophisticated queue/class hierarchies:
- **interactive/serial**: For head jobs and orchestration
- **batch/standard**: Default compute jobs
- **short**: Fast turnaround, short jobs
- **long**: Extended walltime
- **highmem**: High memory nodes
- **gpu**: GPU-equipped nodes
- **mpi**: Parallel/MPI jobs
- **debug**: Quick testing

### Policies and Fairness

Moab provides advanced scheduling features:
- **Fairshare**: Historical usage affects priority
- **Backfill**: Utilizes idle resources efficiently
- **Reservations**: Guaranteed resource allocation
- **QOS**: Quality of Service levels
- **Throttling**: Limits per user/group/account

Be a good citizen:
- Set reasonable `nextflow_queue_size`
- Use appropriate `submitRateLimit`
- Request appropriate walltime
- Use correct queues for job types

## Best Practices

### Resource Management

1. **Set Appropriate Limits**: Use reasonable defaults for memory, CPUs, walltime
2. **Queue Selection**: Use appropriate queues for different job types
3. **Queue Size**: Balance throughput with fair-share policies
4. **Retry Strategy**: Handle transient failures and pre-emption
5. **Walltime**: Request sufficient time but not excessive

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

1. **Parallel Execution**: Leverage Moab's scheduling capabilities
2. **Node Selection**: Use appropriate node types (CPU, GPU, high-mem)
3. **Local Scratch**: Use node-local storage for I/O-intensive tasks
4. **Network**: Ensure good connectivity to shared filesystem
5. **Job Size**: Balance job size with scheduler overhead

### Security

1. **Credentials**: Secure SSH credentials with proper permissions
2. **File Permissions**: Use appropriate umask and permissions
3. **Scratch Cleanup**: Clean sensitive data from scratch
4. **Audit Logs**: Monitor Moab accounting for usage
5. **Access Control**: Use Moab ACLs for multi-tenant environments

### Monitoring

1. **Moab Commands**: Use `showq`, `checkjob`, `showstats` for monitoring
2. **Queue Status**: Check job status and queue information
3. **Resource Usage**: Monitor with Moab accounting tools
4. **Nextflow Reports**: Enable timeline, trace, and report
5. **Filesystem**: Monitor I/O and space usage

## Troubleshooting

### Common Issues

1. **Jobs Not Starting**:
   - Check queue availability: `showq`
   - Verify account/project limits: `showstats`
   - Check job holds: `checkjob JOBID`
   - Review fairshare allocation

2. **Permission Denied**:
   - Verify work directory permissions
   - Check SSH credentials
   - Ensure filesystem is mounted on compute nodes
   - Verify user account status

3. **Module Not Found**:
   - Verify module is available: `module avail`
   - Load required modules in pre-run script
   - Check module dependencies

4. **Singularity Errors**:
   - Verify Singularity is available on compute nodes
   - Check cache directory permissions
   - Ensure bind mounts are correct
   - Verify image format compatibility

5. **Walltime Exceeded**:
   - Increase walltime allocation
   - Use appropriate queue for job length
   - Check job resource usage
   - Split long jobs into stages

6. **Shared Filesystem Issues**:
   - Verify filesystem is mounted
   - Check network connectivity
   - Monitor I/O performance
   - Check quotas and disk space

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
- Resource `launch_directory` → API `config.launchDir`
- Resource `head_queue_name` → API `config.headQueue`
- Resource `compute_queue_name` → API `config.computeQueue`
- Resource `resource_labels` → API `config.resourceLabels`
- Resource `staging_options.pre_run_script` → API `config.preRunScript`
- Resource `staging_options.post_run_script` → API `config.postRunScript`
- Resource `nextflow_config` → API `config.environment` or `config.nextflowConfig`
- Resource `advanced.nextflow_queue_size` → API `config.maxQueueSize`
- Resource `advanced.head_job_submit_options` → API `config.headJobSubmitOptions`
- Resource `advanced.apply_head_job_submit_options` → API `config.propagateHeadJobSubmitOptions`

### Platform Type

- **API Value**: `"moab-platform"`
- **Config Type**: `"MoabComputeConfig"`

## Related Resources

- `seqera_moab_credential` - SSH credentials for Moab cluster head node
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## Moab Cluster Requirements

### Required Software

- **Moab Workload Manager**: 8.x or later
- **PBS/Torque**: Compatible resource manager
- **Nextflow**: **22.10.0 or later** (required)
- **Java**: 11 or later
- **Singularity/Apptainer**: 3.x or later (recommended)
- **Shared Filesystem**: NFS, Lustre, GPFS, etc.

### Network Requirements

- SSH access to Moab submit/login node
- Outbound connectivity to Seqera Cloud web service
- Shared filesystem accessible from all nodes

### User Requirements

- User account on HPC cluster
- SSH key-based authentication
- Read/write access to shared filesystem
- Job submission permissions in Moab
- Appropriate account/project allocation

## References

- [Moab Workload Manager Documentation](https://adaptivecomputing.com/cherry-services/moab-hpc-suite/)
- [Nextflow Executor Documentation](https://www.nextflow.io/docs/latest/executor.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Singularity/Apptainer](https://apptainer.org/)
- [Environment Modules](https://modules.readthedocs.io/)
- [PBS/Torque Documentation](https://www.adaptivecomputing.com/cherry-services/torque-resource-manager/)
