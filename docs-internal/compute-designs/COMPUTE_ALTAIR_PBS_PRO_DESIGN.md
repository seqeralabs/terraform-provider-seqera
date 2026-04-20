# Altair PBS Pro Compute Environment Resource Design

## Overview

This document describes the design for the `seqera_compute_altair_pbs_pro` resource, which manages Altair PBS Pro compute environments in Seqera Platform.

Altair PBS Pro compute environments enable running Nextflow workflows on HPC (High Performance Computing) clusters managed by Altair PBS Professional, providing enterprise-grade scheduling, resource management, and workload optimization for traditional supercomputing infrastructure.

## Key Characteristics

- **Enterprise HPC**: Commercial-grade workload manager for mission-critical systems
- **Advanced Scheduling**: Sophisticated algorithms for resource optimization
- **On-Premises**: Works with on-premises, university, and national lab HPC systems
- **Shared Filesystem**: Requires POSIX-compliant shared filesystem (NFS, Lustre, GPFS, etc.)
- **High Scalability**: Designed for large-scale HPC deployments
- **Policy-Based**: Complex scheduling policies, priorities, and resource limits

## Prerequisites

To connect Seqera Cloud to your local cluster and launch pipeline executions, the following requirements must be fulfilled:

1. **SSH Access**: The cluster must be reachable via an SSH connection using an SSH key
2. **Outbound Connectivity**: The cluster must allow outbound connections to the Seqera Cloud web service
3. **Job Submission**: The cluster queue used to run the Nextflow head job must be able to submit cluster jobs
4. **Nextflow Version**: Nextflow runtime version **22.10.0** (or later) must be pre-installed in the cluster

## Resource Structure

```hcl
resource "seqera_compute_altair_pbs_pro" "example" {
  name         = "pbspro-compute"
  workspace_id = seqera_workspace.main.id

  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id

  # Work directory (must be on shared filesystem)
  work_directory = "/shared/nextflow/work"

  # Optional: Launch directory
  launch_directory = "/shared/nextflow/launch"

  # Queue configuration
  head_queue_name    = "interactive"
  compute_queue_name = "workq"

  # Resource labels
  resource_labels = [
    {
      name  = "seqera_workspace"
      value = "genomics-workspace"
    },
    {
      name  = "environment"
      value = "production"
    }
  ]

  # Staging options
  staging_options {
    pre_run_script  = "#!/bin/bash\nmodule load nextflow java singularity"
    post_run_script = "#!/bin/bash\necho 'Workflow complete'"
  }

  # Nextflow configuration
  nextflow_config = <<-EOF
    process {
      executor = 'pbspro'
      queue = 'workq'
    }
  EOF

  # Environment variables
  environment_variables = {
    "PBS_SERVER" = "pbspro-head.example.com"
  }

  # Advanced options
  advanced {
    nextflow_queue_size              = 100
    head_job_submit_options          = "-l walltime=24:00:00,mem=8gb"
    apply_head_job_submit_options    = false
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
- **Example**: `"pbspro-hpc"`, `"altair-cluster"`

#### `workspace_id`
- **Type**: Integer (Int64)
- **Required**: Yes (Optional for user context)
- **Description**: Workspace numeric identifier where the compute environment will be created
- **Example**: `123456`

#### `credentials_id`
- **Type**: String
- **Required**: Yes
- **Description**: Altair PBS Pro credentials ID to use for accessing the HPC cluster
- **Reference**: Must reference a valid `seqera_altair_pbs_pro_credential` resource
- **Example**: `seqera_altair_pbs_pro_credential.main.credentials_id`
- **Options**: Can be SSH-based or Tower agent-based credentials
- **Notes**: Credentials contain SSH connection details for the PBS Pro cluster

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
- **Example**: `"interactive"`, `"serial"`, `"submit"`, `"express"`
- **Character Limit**: 0/100 characters
- **Notes**:
  - Queue for the Nextflow head/orchestrator job
  - Should support long-running jobs
  - Must have job submission capabilities

#### `compute_queue_name`
- **Type**: String
- **Optional**: Yes
- **Description**: The name of queue on the cluster to which pipeline jobs are submitted
- **Example**: `"workq"`, `"batch"`, `"compute"`, `"normal"`
- **Character Limit**: 0/100 characters
- **Default**: PBS Pro default is typically `"workq"`
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
      value = "pbspro-production"
    },
    {
      name  = "cost_center"
      value = "research"
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

  # Check PBS Pro server connectivity
  qstat -B > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "ERROR: Cannot connect to PBS Pro server"
    exit 1
  fi

  echo "PBS Pro environment initialized"
  ```
- **Use Cases**:
  - Load environment modules
  - Set up temporary directories
  - Validate filesystem access
  - Configure Singularity/Apptainer
  - Verify PBS Pro connectivity
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

  # Query job statistics
  for jobid in $(qstat -u $USER -x | grep -v "Job id" | awk '{print $1}'); do
    qstat -f $jobid >> /shared/nextflow/job_stats_$TIMESTAMP.txt
  done

  # Send notification
  echo "Workflow completed at $TIMESTAMP with status $NXF_EXIT_STATUS" | \
    mail -s "Pipeline Complete" user@example.com
  ```
- **Use Cases**:
  - Archive results to long-term storage
  - Cleanup temporary files
  - Collect PBS Pro job statistics
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
    executor = 'pbspro'
    queue = 'workq'

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
      clusterOptions = '-l select=1:ncpus=4:ngpus=1:mem=32gb'
    }

    withLabel: mpi {
      queue = 'mpi'
      clusterOptions = '-l select=4:ncpus=16:mpiprocs=16:mem=64gb'
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
  - Configure PBS Pro executor settings
  - Set default resource requirements
  - Define queue mappings
  - Configure Singularity/Apptainer
  - Set retry strategies
  - Define per-label resource requirements

### Environment Variables

#### `environment_variables`
- **Type**: Map of String to String
- **Optional**: Yes
- **Description**: Environment variables set in all workflow jobs
- **Example**:
  ```hcl
  environment_variables = {
    "PBS_SERVER"                = "pbspro-head.example.com"
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
Advanced configuration options for PBS Pro integration.

##### `advanced.nextflow_queue_size`
- **Type**: Integer
- **Optional**: Yes
- **Description**: The maximum number of jobs Nextflow can submit to the queue simultaneously
- **Default**: `100`
- **Range**: 1 to thousands (depends on cluster limits and policies)
- **Example**: `500`
- **Notes**:
  - Controls how many PBS Pro jobs Nextflow submits simultaneously
  - Should consider cluster fair-share policies
  - Higher values for large parallel workflows
  - Lower values to be a good citizen on shared clusters

##### `advanced.head_job_submit_options`
- **Type**: String
- **Optional**: Yes
- **Description**: Grid Engine submit options for the Nextflow head job
- **Format**: PBS Pro options (qsub parameters) without the `qsub` command
- **Examples**:
  - `"-l walltime=24:00:00,mem=8gb"`
  - `"-l select=1:ncpus=4:mem=16gb -l walltime=48:00:00"`
  - `"-l walltime=72:00:00 -q long -P project123"`
- **Character Limit**: 0/200 characters
- **Notes**:
  - Options are added to the 'qsub' command run by Seqera Cloud to launch the pipeline execution
  - Common PBS Pro options:
    - `-l walltime=HH:MM:SS`: Maximum runtime
    - `-l select=N:ncpus=C:mem=Mgb`: Resource selection
    - `-q queue`: Queue name
    - `-P project`: Project/account code
    - `-l place=`: Placement directive (e.g., `scatter`, `pack`)
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
4. **head_job_submit_options**: Must be valid PBS Pro options format
5. **resource_labels**: Each label must have both `name` and `value`

### Lifecycle Considerations

- **Create**: Configures PBS Pro compute environment in Seqera Platform
- **Read**: Retrieves current state from Seqera Platform API
- **Update**: Updates mutable fields
- **Delete**: Removes compute environment (doesn't affect PBS Pro cluster)

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
resource "seqera_compute_altair_pbs_pro" "minimal" {
  name           = "pbspro-minimal"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory = "/shared/nextflow/work"
}
```

### Standard Configuration

```hcl
resource "seqera_compute_altair_pbs_pro" "standard" {
  name           = "pbspro-standard"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory = "/gpfs/project/nextflow/work"
  launch_directory = "/gpfs/project/nextflow/launch"

  head_queue_name    = "interactive"
  compute_queue_name = "workq"

  resource_labels = [
    {
      name  = "cluster"
      value = "pbspro-hpc"
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
    head_job_submit_options = "-l walltime=24:00:00,select=1:ncpus=2:mem=8gb"
  }
}
```

### Production with Multiple Queues

```hcl
resource "seqera_compute_altair_pbs_pro" "production" {
  name           = "pbspro-prod"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory = "/lustre/scratch/nextflow/work"
  launch_directory = "/lustre/scratch/nextflow/launch"

  head_queue_name    = "express"
  compute_queue_name = "workq"

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

      # Verify PBS Pro connectivity
      qstat -B > /dev/null 2>&1
      if [ $? -ne 0 ]; then
        echo "ERROR: Cannot connect to PBS Pro server"
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

      # Collect job statistics
      for jobid in $(qstat -u $USER -x | grep -v "Job id" | awk '{print $1}'); do
        qstat -f $jobid >> $ARCHIVE_DIR/pbs_job_stats.txt
      done

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
      executor = 'pbspro'
      queue = 'workq'

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
        clusterOptions = '-l select=1:ncpus=8:mem=128gb -l walltime=24:00:00'
      }

      // GPU processes
      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '-l select=1:ncpus=4:ngpus=1:mem=32gb -l walltime=8:00:00'
      }

      // MPI processes
      withLabel: mpi {
        queue = 'mpi'
        clusterOptions = '-l select=4:ncpus=16:mpiprocs=16:mem=64gb -l walltime=12:00:00'
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
    "PBS_SERVER"                = "pbspro-head.example.com"
    "NXF_ANSI_LOG"              = "false"
    "NXF_OPTS"                  = "-Xms2g -Xmx8g"
    "TMPDIR"                    = "/scratch/$USER/tmp"
    "NXF_SINGULARITY_CACHEDIR"  = "/shared/singularity/cache"
  }

  advanced {
    nextflow_queue_size         = 200
    head_job_submit_options     = "-l walltime=48:00:00,select=1:ncpus=4:mem=16gb -q express"
    apply_head_job_submit_options = false
  }
}
```

### GPU Workloads

```hcl
resource "seqera_compute_altair_pbs_pro" "gpu" {
  name           = "pbspro-gpu"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory = "/gpfs/scratch/gpu-work"

  head_queue_name    = "express"
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
      executor = 'pbspro'
      queue = 'gpu'

      withLabel: gpu {
        queue = 'gpu'
        clusterOptions = '-l select=1:ncpus=4:ngpus=2:mem=64gb -l walltime=24:00:00'
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
resource "seqera_compute_altair_pbs_pro" "mpi" {
  name           = "pbspro-mpi"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory = "/lustre/project/mpi-work"

  head_queue_name    = "express"
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
      executor = 'pbspro'
      queue = 'mpi'

      withLabel: mpi {
        queue = 'mpi'
        clusterOptions = '-l select=8:ncpus=16:mpiprocs=16:mem=64gb -l walltime=12:00:00 -l place=scatter'
      }
    }
  EOF
}
```

### High-Throughput Workloads

```hcl
resource "seqera_compute_altair_pbs_pro" "high_throughput" {
  name           = "pbspro-highthroughput"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory = "/gpfs/scratch/high-throughput"

  head_queue_name    = "express"
  compute_queue_name = "short"

  resource_labels = [
    {
      name  = "workload_type"
      value = "high_throughput"
    }
  ]

  nextflow_config = <<-EOF
    process {
      executor = 'pbspro'
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

### National Lab HPC Cluster

```hcl
resource "seqera_compute_altair_pbs_pro" "national_lab" {
  name           = "pbspro-national-lab"
  workspace_id   = seqera_workspace.main.id
  credentials_id = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory = "/lustre/project/nextflow/work"
  launch_directory = "/lustre/project/nextflow/launch"

  head_queue_name    = "regular"
  compute_queue_name = "batch"

  resource_labels = [
    {
      name  = "institution"
      value = "national-lab"
    },
    {
      name  = "project"
      value = "genomics"
    }
  ]

  staging_options {
    pre_run_script = <<-EOF
      #!/bin/bash
      # Load standard modules
      module load nextflow java singularity

      # Set up project-specific directories
      export PROJECT_DIR=/lustre/project/genomics
      export TMPDIR=/scratch/$USER/tmp
      mkdir -p $TMPDIR

      # Validate project allocation
      qstat -Qf batch | grep -q "project_genomics"
      if [ $? -ne 0 ]; then
        echo "ERROR: Project allocation not found"
        exit 1
      fi
    EOF

    post_run_script = <<-EOF
      #!/bin/bash
      # Archive to long-term storage
      TIMESTAMP=$(date +%Y%m%d_%H%M%S)
      hsi "cd /archive/genomics; mkdir $TIMESTAMP; put -R /lustre/project/nextflow/results/* $TIMESTAMP/"

      # Cleanup old work files (>30 days)
      find /lustre/project/nextflow/work -type f -mtime +30 -delete
    EOF
  }

  nextflow_config = <<-EOF
    process {
      executor = 'pbspro'
      queue = 'batch'

      // Default resources for national lab
      cpus = 8
      memory = '32 GB'
      time = '4h'

      errorStrategy = 'retry'
      maxRetries = 2

      // Project code for accounting
      clusterOptions = '-P genomics'

      withLabel: large_scale {
        cpus = 32
        memory = '128 GB'
        time = '24h'
        clusterOptions = '-l select=1:ncpus=32:mem=128gb -l walltime=24:00:00 -P genomics'
      }
    }

    executor {
      queueSize = 100
      pollInterval = '1 min'
      queueStatInterval = '10 min'
    }

    singularity {
      enabled = true
      autoMounts = true
      cacheDir = '/lustre/project/singularity/cache'
    }
  EOF

  advanced {
    nextflow_queue_size     = 100
    head_job_submit_options = "-l walltime=48:00:00,select=1:ncpus=4:mem=16gb -P genomics"
  }
}
```

## PBS Pro-Specific Considerations

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

### PBS Pro Resource Specifications

PBS Pro uses a sophisticated resource selection syntax:

**Select Statement Format**:
```bash
-l select=N:ncpus=C:mem=Mgb[:ngpus=G][:mpiprocs=M]
```

**Common Resource Specifications**:
- **Walltime**: `-l walltime=HH:MM:SS` or `-l walltime=D:HH:MM:SS`
- **Memory**: Specified in select statement: `select=1:ncpus=4:mem=16gb`
- **CPUs**: Specified in select statement: `select=1:ncpus=8`
- **GPUs**: Specified in select statement: `select=1:ncpus=4:ngpus=2:mem=32gb`
- **MPI**: Specified with mpiprocs: `select=4:ncpus=16:mpiprocs=16:mem=64gb`
- **Queue**: `-q queuename`
- **Project/Account**: `-P projectname`
- **Placement**: `-l place=scatter|pack|free`

**Examples**:
```bash
# Single node, 8 CPUs, 32 GB memory, 24 hour walltime
-l select=1:ncpus=8:mem=32gb -l walltime=24:00:00

# GPU job with 4 CPUs and 2 GPUs
-l select=1:ncpus=4:ngpus=2:mem=64gb -l walltime=8:00:00

# MPI job on 8 nodes, 16 CPUs per node
-l select=8:ncpus=16:mpiprocs=16:mem=64gb -l walltime=12:00:00 -l place=scatter
```

### PBS Pro ClusterOptions

Per-process PBS Pro options in Nextflow config:
```groovy
process {
  withLabel: big_mem {
    clusterOptions = '-l select=1:ncpus=8:mem=128gb -l walltime=24:00:00'
  }

  withLabel: gpu {
    clusterOptions = '-l select=1:ncpus=4:ngpus=1:mem=32gb -l walltime=8:00:00 -q gpu'
  }

  withLabel: mpi {
    clusterOptions = '-l select=4:ncpus=16:mpiprocs=16:mem=64gb -l walltime=12:00:00 -l place=scatter'
  }
}
```

### Job Queues

PBS Pro supports multiple queues with different characteristics:
- **workq**: Default execution queue
- **express/interactive**: For head jobs and quick tasks
- **batch**: Standard compute jobs
- **short**: Fast turnaround, short jobs
- **long**: Extended walltime
- **highmem**: High memory nodes
- **gpu**: GPU-equipped nodes
- **mpi**: Parallel/MPI jobs
- **debug**: Quick testing

### Scheduling Policies

PBS Pro provides enterprise-grade scheduling features:
- **Fairshare**: Historical usage affects priority
- **Backfill**: Utilizes idle resources efficiently
- **Reservations**: Guaranteed resource allocation
- **Job Arrays**: Efficient handling of similar jobs
- **Preemption**: Priority-based job preemption
- **Resource Limits**: Per-user, per-group, per-project limits

Be a good citizen:
- Set reasonable `nextflow_queue_size`
- Use appropriate `submitRateLimit`
- Request appropriate walltime
- Use correct queues for job types
- Respect project allocations

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

1. **Parallel Execution**: Leverage PBS Pro's scheduling capabilities
2. **Node Selection**: Use appropriate node types (CPU, GPU, high-mem)
3. **Local Scratch**: Use node-local storage for I/O-intensive tasks
4. **Network**: Ensure good connectivity to shared filesystem
5. **Job Size**: Balance job size with scheduler overhead

### Security

1. **Credentials**: Secure SSH credentials with proper permissions
2. **File Permissions**: Use appropriate umask and permissions
3. **Scratch Cleanup**: Clean sensitive data from scratch
4. **Audit Logs**: Monitor PBS Pro accounting for usage
5. **Access Control**: Use PBS Pro ACLs for multi-tenant environments

### Monitoring

1. **PBS Pro Commands**: Use `qstat`, `qstat -f`, `qstat -B` for monitoring
2. **Queue Status**: Check job status and queue information
3. **Resource Usage**: Monitor with PBS Pro accounting tools (`qstat -x`, `tracejob`)
4. **Nextflow Reports**: Enable timeline, trace, and report
5. **Filesystem**: Monitor I/O and space usage

## Troubleshooting

### Common Issues

1. **Jobs Not Starting**:
   - Check queue availability: `qstat -Q`
   - Verify project/account limits: `qstat -Qf queuename`
   - Check job status: `qstat -f jobid`
   - Review server status: `qstat -B`

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
   - Check job resource usage with `qstat -f`
   - Split long jobs into stages

6. **Shared Filesystem Issues**:
   - Verify filesystem is mounted
   - Check network connectivity
   - Monitor I/O performance
   - Check quotas and disk space

7. **PBS Pro Server Connection**:
   - Verify PBS_SERVER environment variable
   - Check network connectivity: `qstat -B`
   - Validate credentials
   - Check PBS Pro service status on server

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

- **API Value**: `"altair-pbspro-platform"` or `"pbspro-platform"`
- **Config Type**: `"PbsProComputeConfig"`

## Related Resources

- `seqera_altair_pbs_pro_credential` - SSH credentials for PBS Pro cluster head node
- `seqera_workspace` - Workspace containing the compute environment
- `seqera_primary_compute_env` - Set this as the primary compute environment

## PBS Pro Cluster Requirements

### Required Software

- **Altair PBS Professional**: 19.x or later
- **Nextflow**: **22.10.0 or later** (required)
- **Java**: 11 or later
- **Singularity/Apptainer**: 3.x or later (recommended)
- **Shared Filesystem**: NFS, Lustre, GPFS, etc.

### Network Requirements

- SSH access to PBS Pro submit/login node
- Outbound connectivity to Seqera Cloud web service
- Shared filesystem accessible from all nodes

### User Requirements

- User account on HPC cluster
- SSH key-based authentication
- Read/write access to shared filesystem
- Job submission permissions in PBS Pro
- Appropriate project/account allocation

## References

- [Altair PBS Professional Documentation](https://www.altair.com/pbs-professional/)
- [PBS Pro User Guide](https://help.altair.com/2022.1.0/PBS%20Professional/PBSUserGuide2022.1.pdf)
- [Nextflow Executor Documentation](https://www.nextflow.io/docs/latest/executor.html)
- [Seqera Platform Documentation](https://docs.seqera.io/)
- [Singularity/Apptainer](https://apptainer.org/)
- [Environment Modules](https://modules.readthedocs.io/)
