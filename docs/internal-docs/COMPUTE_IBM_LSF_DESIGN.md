# IBM LSF Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_ibm_lsf` resource, which manages IBM LSF (Load Sharing Facility) compute environments in Seqera Platform.

IBM LSF compute environments enable running Nextflow workflows on HPC (High Performance Computing) clusters managed by IBM LSF, providing enterprise-grade workload management, advanced scheduling, and resource optimization for traditional supercomputing infrastructure.

## Key Characteristics

- **Enterprise HPC**: Designed for large-scale enterprise and academic HPC clusters
- **Advanced Scheduling**: Sophisticated job scheduling with priorities, fairshare, and policies
- **On-Premises**: Works with on-premises and cloud-deployed LSF clusters
- **Shared Filesystem**: Requires POSIX-compliant shared filesystem (NFS, Lustre, GPFS, etc.)
- **Resource Optimization**: Advanced resource management and job preemption
- **Multi-Tenant**: Support for complex multi-tenant environments

## Prerequisites

To connect Seqera Cloud to your local cluster and launch pipeline executions, the following requirements must be fulfilled:

1. **SSH Access**: The cluster must be reachable via an SSH connection using an SSH key
2. **Outbound Connectivity**: The cluster must allow outbound connections to the Seqera Cloud web service
3. **Job Submission**: The cluster queue used to run the Nextflow head job must be able to submit cluster jobs
4. **Nextflow Version**: Nextflow runtime version **22.10.0** (or later) must be pre-installed in the cluster

## Resource Structure

```hcl
resource "seqera_compute_ibm_lsf" "example" {
  name         = "lsf-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_lsf_credential.main.credentials_id

  # Work directory (must be on shared filesystem)
  work_directory = "/shared/nextflow/work"

  # Optional: Launch directory
  launch_directory = "/shared/nextflow/launch"

  # Queue configuration
  head_queue_name    = "interactive"
  compute_queue_name = "normal"

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
      executor = 'lsf'
      queue = 'normal'
    }
  EOF

  # Environment variables
  environment_variables = {
    "LSF_CLUSTER" = "hpc-cluster"
  }

  # Advanced options
  advanced {
    nextflow_queue_size               = 100
    head_job_submit_options           = "-W 24:00 -M 8GB"
    apply_head_job_submit_options     = false
    unit_for_memory_limits            = "MB"
    per_job_memory_limit              = true
    per_task_reserve                  = false
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
- **Example**: `"lsf-hpc"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: IBM LSF credentials ID to use for accessing the HPC cluster
- **Reference**: Must reference a valid `seqera_lsf_credential` resource
- **Example**: `seqera_lsf_credential.main.credentials_id`
- **Notes**: Credentials contain SSH connection details for the LSF cluster

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
- **Example**: `"interactive"`, `"serial"`, `"short"`, `"priority"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Queue for the Nextflow head/orchestrator job
  - Should support long-running jobs
  - Must have job submission capabilities

#### `compute_queue_name`
- **Type**: String
- **Optional**: Yes
- **Description**: The name of queue on the cluster to which pipeline jobs are submitted
- **Example**: `"normal"`, `"batch"`, `"compute"`, `"standard"`
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
      value = "lsf-prod"
    },
    {
      name  = "queue"
      value = "normal"
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

  echo "LSF environment initialized"
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
    executor = 'lsf'
    queue = 'normal'

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
      clusterOptions = '-gpu "num=1:mode=shared:mps=no"'
    }

    withLabel: mpi {
      queue = 'parallel'
      clusterOptions = '-n 32 -R "span[ptile=16]"'
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
  - Configure LSF executor settings
  - Set default resource requirements
  - Define queue mappings
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
    "LSF_CLUSTER"               = "hpc-cluster"
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
Advanced configuration options for LSF integration.

##### `advanced.nextflow_queue_size`
- **Type**: Integer
- **Optional**: Yes
- **Description**: The maximum number of jobs Nextflow can submit to the queue simultaneously
- **Default**: `100`
- **Range**: 1 to thousands (depends on cluster limits)
- **Example**: `500`
- **Notes**:
  - Controls how many LSF jobs Nextflow submits simultaneously
  - Should consider cluster fair-share policies
  - Higher values for large parallel workflows
  - Lower values to be a good citizen on shared clusters

##### `advanced.head_job_submit_options`
- **Type**: String
- **Optional**: Yes
- **Description**: Grid Engine submit options for the Nextflow head job
- **Format**: LSF bsub options without the `bsub` command
- **Examples**:
  - `"-W 24:00 -M 8GB"`
  - `"-W 48:00 -M 16GB -n 4"`
  - `"-W 72:00 -M 32GB -q long -P project123"`
- **Character Limit**: 0/200 characters
- **Notes**:
  - These options are added to the 'qsub' command run by Seqera Cloud to launch the pipeline execution
  - Common options: `-W` (walltime), `-M` (memory limit), `-n` (cores), `-q` (queue), `-P` (project)
  - Head job runs for entire workflow duration
  - Should request sufficient time for workflow completion

##### `advanced.apply_head_job_submit_options`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Enable this field so the "Head job submit options" to also be applied to the Nextflow-submitted compute jobs
- **Default**: `false`
- **Example**: `true`
- **Notes**:
  - If true, options from `head_job_submit_options` are applied to all jobs
  - Usually false, as compute jobs need different resources
  - Use `clusterOptions` in Nextflow config for per-process options

##### `advanced.unit_for_memory_limits`
- **Type**: String (Dropdown)
- **Optional**: Yes
- **Description**: Define the unit used by your LSF cluster for memory limits
- **Allowed Values**:
  - `"MB"` - Megabytes
  - `"GB"` - Gigabytes
  - `"KB"` - Kilobytes
- **Default**: Varies by cluster configuration
- **Example**: `"MB"`
- **Notes**:
  - This must match the LSF_UNIT_FOR_LIMITS value in your lsf.conf file
  - Critical for correct memory allocation
  - Mismatch can cause job failures or resource issues

##### `advanced.per_job_memory_limit`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Define whether the memory limit is interpreted as per-job or per-process
- **Default**: `false` (varies by cluster)
- **Example**: `true`
- **Notes**:
  - This must match the LSB_JOB_MEMLIMIT value in your lsf.conf file
  - `true`: Memory limit applies to entire job
  - `false`: Memory limit applies per process/task
  - Important for multi-core jobs

##### `advanced.per_task_reserve`
- **Type**: Boolean
- **Optional**: Yes
- **Description**: Define whether the memory reservation is made on job tasks or per-task
- **Default**: `false` (varies by cluster)
- **Example**: `true`
- **Notes**:
  - This must match the RESOURCE_RESERVE_PER_TASK value in your lsf.conf file
  - Affects how memory reservations are calculated
  - Important for resource scheduling accuracy

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
4. **head_job_submit_options**: Must be valid LSF options format
5. **unit_for_memory_limits**: Must be one of: MB, GB, KB
6. **resource_labels**: Each label must have both `name` and `value`

### Lifecycle Considerations

- **Create**: Configures IBM LSF compute environment in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields
- **Delete**: Removes compute environment (doesn't affect LSF cluster)

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
resource "seqera_compute_ibm_lsf" "minimal" {
  name           = "lsf-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_lsf_credential.main.credentials_id
  work_directory = "/shared/nextflow/work"
}
```

### Standard Configuration

```hcl
resource "seqera_compute_ibm_lsf" "standard" {
  name           = "lsf-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_lsf_credential.main.credentials_id
  work_directory = "/gpfs/project/nextflow/work"
  launch_directory = "/gpfs/project/nextflow/launch"

  head_queue_name    = "interactive"
  compute_queue_name = "normal"

  resource_labels = [
    {
      name  = "cluster"
      value = "lsf-hpc"
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
    head_job_submit_options = "-W 24:00 -M 8GB"
    unit_for_memory_limits = "MB"
    per_job_memory_limit = true
  }
}
```

### Production with Multiple Queues

```hcl
resource "seqera_compute_ibm_lsf" "production" {
  name           = "lsf-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_lsf_credential.main.credentials_id
  work_directory = "/lustre/scratch/nextflow/work"
  launch_directory = "/lustre/scratch/nextflow/launch"

  head_queue_name    = "serial"
  compute_queue_name = "normal"

  resource_labels = [
    {
      name  = "environment"
      value = "production"
    },
    {
      name  = "cluster"
      value = "enterprise-lsf"
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

      echo "LSF production environment initialized"
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

      # Send notification via LSF
      bsub -q notification -J notify "echo 'Workflow completed. Results: $ARCHIVE_DIR' | mail -s 'Pipeline Complete' user@example.com"
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'lsf'
      queue = 'normal'

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
        clusterOptions = '-M 128000 -R "rusage[mem=128000]"'
      }

      // GPU processes
      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '-gpu "num=1:mode=shared:mps=no" -R "rusage[ngpus_physical=1.00]"'
        time = '8h'
      }

      // Parallel/MPI processes
      withLabel: mpi {
        queue = 'parallel'
        clusterOptions = '-n 32 -R "span[ptile=16]"'
        time = '12h'
      }

      // Fast/short jobs
      withLabel: short {
        queue = 'short'
        time = '30m'
      }

      // Priority jobs
      withLabel: priority {
        queue = 'priority'
        clusterOptions = '-sp 100'
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
    "LSF_CLUSTER"               = "enterprise-lsf"
    "NXF_ANSI_LOG"              = "false"
    "NXF_OPTS"                  = "-Xms2g -Xmx8g"
    "TMPDIR"                    = "/scratch/$USER/tmp"
    "NXF_SINGULARITY_CACHEDIR"  = "/shared/singularity/cache"
  }

  advanced {
    nextflow_queue_size           = 200
    head_job_submit_options       = "-W 48:00 -M 16GB -n 4 -q serial -P genomics"
    apply_head_job_submit_options = false
    unit_for_memory_limits        = "MB"
    per_job_memory_limit          = true
    per_task_reserve              = false
  }
}
```

### GPU Workloads

```hcl
resource "seqera_compute_ibm_lsf" "gpu" {
  name           = "lsf-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_lsf_credential.main.credentials_id
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
      executor = 'lsf'
      queue = 'gpu'

      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '-gpu "num=2:mode=shared:mps=no:j_exclusive=yes" -R "rusage[ngpus_physical=2.00]"'
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
    unit_for_memory_limits = "MB"
  }
}
```

### Parallel/MPI Workloads

```hcl
resource "seqera_compute_ibm_lsf" "mpi" {
  name           = "lsf-mpi"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_lsf_credential.main.credentials_id
  work_directory = "/lustre/project/mpi-work"

  head_queue_name    = "serial"
  compute_queue_name = "parallel"

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
      executor = 'lsf'
      queue = 'parallel'

      withLabel: mpi {
        queue = 'parallel'
        clusterOptions = '-n 128 -R "span[ptile=16]"'
      }
    }
  EOF

  advanced {
    unit_for_memory_limits = "GB"
    per_job_memory_limit = true
  }
}
```

### High-Throughput Workloads

```hcl
resource "seqera_compute_ibm_lsf" "high_throughput" {
  name           = "lsf-highthroughput"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_lsf_credential.main.credentials_id
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
      executor = 'lsf'
      queue = 'short'

      errorStrategy = { task.exitStatus in [130,140] ? 'retry' : 'ignore' }
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
    unit_for_memory_limits = "MB"
  }
}
```

## IBM LSF-Specific Considerations

### Shared Filesystem Requirements

The work directory **must** be on a shared filesystem accessible from:
- The LSF master/submit host (where Nextflow runs)
- All LSF execution hosts (where jobs execute)

Common shared filesystems:
- **GPFS/Spectrum Scale**: IBM high-performance filesystem (often used with LSF)
- **Lustre**: High-performance parallel filesystem
- **NFS**: Simple, widely used
- **BeeGFS**: Parallel cluster filesystem
- **CephFS**: Distributed filesystem

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
```groovy
singularity {
  enabled = true
  autoMounts = true
  cacheDir = '/shared/singularity/cache'
  runOptions = '--bind /gpfs:/gpfs'
}
```

### LSF Resource Specifications

IBM LSF has specific resource syntax:
- **Walltime**: `-W HH:MM` (hours:minutes)
- **Memory Limit**: `-M <value>` (in MB, GB, or KB based on LSF_UNIT_FOR_LIMITS)
- **Memory Reservation**: `-R "rusage[mem=<value>]"`
- **Cores**: `-n <count>`
- **GPUs**: `-gpu "num=<count>:mode=<mode>"`
- **Span**: `-R "span[ptile=<count>]"` (processes per host)
- **Queue**: `-q <queuename>`
- **Project**: `-P <projectname>`
- **Priority**: `-sp <value>`

### LSF ClusterOptions

Per-process LSF options in Nextflow config:
```groovy
process {
  withLabel: big_mem {
    clusterOptions = '-M 128000 -R "rusage[mem=128000]" -W 24:00'
  }

  withLabel: gpu {
    clusterOptions = '-gpu "num=1:mode=shared" -R "rusage[ngpus_physical=1.00]"'
  }

  withLabel: mpi {
    clusterOptions = '-n 32 -R "span[ptile=16]"'
  }

  withLabel: priority {
    clusterOptions = '-sp 100 -P urgent'
  }
}
```

### Memory Configuration

**Critical**: LSF memory configuration must match cluster settings:

1. **unit_for_memory_limits**: Must match `LSF_UNIT_FOR_LIMITS` in lsf.conf
2. **per_job_memory_limit**: Must match `LSB_JOB_MEMLIMIT` in lsf.conf
3. **per_task_reserve**: Must match `RESOURCE_RESERVE_PER_TASK` in lsf.conf

Mismatch can cause:
- Job rejection
- Insufficient memory allocation
- Incorrect resource reservation

### Queues and Job Classes

LSF supports sophisticated queue hierarchies:
- **interactive/serial**: For head jobs and orchestration
- **normal/batch**: Default compute jobs
- **short**: Fast turnaround, short jobs
- **long**: Extended walltime
- **highmem**: High memory hosts
- **gpu**: GPU-equipped hosts
- **parallel**: Parallel/MPI jobs
- **priority**: High-priority execution

### Resource Limits and Fairshare

LSF provides enterprise-grade features:
- **Fairshare**: Historical usage affects priority
- **Job Slots**: Limit concurrent jobs per user/group
- **Resource Reservation**: Advance reservation of resources
- **Preemption**: Higher priority jobs can preempt lower priority
- **Service Classes**: Different QoS levels
- **Job Arrays**: Efficient submission of parameter sweeps

## Best Practices

### Resource Management

1. **Set Appropriate Limits**: Use reasonable defaults for memory, CPUs, walltime
2. **Queue Selection**: Use appropriate queues for different job types
3. **Queue Size**: Balance throughput with cluster policies
4. **Retry Strategy**: Handle job preemption and failures
5. **Memory Units**: Ensure correct unit configuration

### Shared Filesystem

1. **Work Directory**: Use high-performance shared filesystem (GPFS preferred)
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

1. **Parallel Execution**: Leverage LSF's advanced scheduling
2. **Resource Selection**: Use appropriate resources (CPU, GPU, high-mem)
3. **Local Scratch**: Use host-local storage for I/O-intensive tasks
4. **Network**: Ensure good connectivity to shared filesystem
5. **Job Packing**: Use span and resource requirements efficiently

### Security

1. **Credentials**: Secure SSH credentials with proper permissions
2. **File Permissions**: Use appropriate umask and permissions
3. **Scratch Cleanup**: Clean sensitive data from scratch
4. **Audit Logs**: Monitor LSF accounting for usage
5. **Access Control**: Use LSF ACLs for multi-tenant environments

### Monitoring

1. **LSF Commands**: Use `bjobs`, `bhist`, `bacct` for monitoring
2. **Queue Status**: Check `bqueues` for queue information
3. **Resource Usage**: Monitor with LSF accounting tools
4. **Nextflow Reports**: Enable timeline, trace, and report
5. **Filesystem**: Monitor I/O and space usage

## Troubleshooting

### Common Issues

1. **Jobs Not Starting**:
   - Check queue status: `bqueues`
   - Verify job details: `bjobs -l <jobid>`
   - Check user limits: `bugroup`
   - Review cluster load: `bhosts`

2. **Memory Errors**:
   - Verify `unit_for_memory_limits` matches cluster config
   - Check `per_job_memory_limit` setting
   - Review job memory usage: `bjobs -l <jobid>`
   - Increase memory allocation

3. **Permission Denied**:
   - Verify work directory permissions
   - Check SSH credentials
   - Ensure filesystem is mounted on execution hosts
   - Verify user account status

4. **Module Not Found**:
   - Verify module is available: `module avail`
   - Load required modules in pre-run script
   - Check module dependencies

5. **Singularity Errors**:
   - Verify Singularity is available on execution hosts
   - Check cache directory permissions
   - Ensure bind mounts are correct
   - Verify image format compatibility

6. **Job Killed (Exit 130, 140)**:
   - LSF killed job (resource limit exceeded or walltime)
   - Increase resource requests
   - Check job exit status: `bhist -l <jobid>`
   - Review error logs

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
- Resource `advanced.unit_for_memory_limits` → API `config.unitForMemoryLimits`
- Resource `advanced.per_job_memory_limit` → API `config.perJobMemoryLimit`
- Resource `advanced.per_task_reserve` → API `config.perTaskReserve`

### Platform Type

- **API Value**: `"lsf-platform"` or `"ibm-lsf"`
- **Config Type**: `"LSFComputeConfig"` or `"IBMLSFComputeConfig"`

## Related Resources

- `seqera_lsf_credential` - SSH credentials for LSF cluster
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## IBM LSF Cluster Requirements

### Required Software

- **IBM LSF**: 9.x or later (10.x recommended)
- **Nextflow**: **22.10.0 or later** (required)
- **Java**: 11 or later
- **Singularity/Apptainer**: 3.x or later (recommended)
- **Shared Filesystem**: GPFS, Lustre, NFS, etc.

### Network Requirements

- SSH access to LSF master/submit host
- Outbound connectivity to Seqera Cloud web service
- Shared filesystem accessible from all hosts

### User Requirements

- User account on LSF cluster
- SSH key-based authentication
- Read/write access to shared filesystem
- Job submission permissions in LSF
- Appropriate project/group allocation

## References

- [IBM Spectrum LSF Documentation](https://www.ibm.com/docs/en/spectrum-lsf)
- [Nextflow LSF Executor](https://www.nextflow.io/docs/latest/executor.html#lsf)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Singularity/Apptainer](https://apptainer.org/)
- [Environment Modules](https://modules.readthedocs.io/)
- [IBM Spectrum Scale (GPFS)](https://www.ibm.com/docs/en/spectrum-scale)
