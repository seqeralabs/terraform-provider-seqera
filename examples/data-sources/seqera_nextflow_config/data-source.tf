# Nextflow Config Data Source Examples

# Example 1: Basic local configuration
data "seqera_nextflow_config" "local" {
  process {
    executor = "local"
    cpus     = 2
    memory   = "4 GB"
  }
}

output "local_config" {
  value = data.seqera_nextflow_config.local.config
}

# Example 2: HPC with PBS Pro and multiple labels
data "seqera_nextflow_config" "pbspro" {
  process {
    executor       = "pbspro"
    queue          = "workq"
    cpus           = 4
    memory         = "16 GB"
    time           = "4h"
    error_strategy = "retry"
    max_retries    = 2

    # High memory processes
    with_label {
      name            = "big_mem"
      memory          = "128 GB"
      time            = "24h"
      queue           = "highmem"
      cluster_options = "-l select=1:ncpus=8:mem=128gb"
    }

    # GPU processes
    with_label {
      name            = "gpu"
      queue           = "gpu"
      cluster_options = "-l select=1:ncpus=4:ngpus=2"
    }

    # MPI processes
    with_label {
      name            = "mpi"
      queue           = "mpi"
      cluster_options = "-l select=4:ncpus=16:mpiprocs=16"
    }
  }

  executor {
    queue_size          = 100
    poll_interval       = "30 sec"
    submit_rate_limit   = "10 sec"
  }

  singularity {
    enabled     = true
    auto_mounts = true
    cache_dir   = "/shared/singularity/cache"
    run_options = "--bind /lustre:/lustre"
  }
}

# Example 3: Slurm with process name patterns
data "seqera_nextflow_config" "slurm" {
  process {
    executor = "slurm"
    queue    = "compute"

    # Override for alignment processes
    with_name {
      pattern = "align_*"
      cpus    = 16
      memory  = "64 GB"
      time    = "12h"
    }

    # Override for STAR aligner specifically
    with_name {
      pattern         = "STAR"
      cpus            = 32
      memory          = "128 GB"
      cluster_options = "--gres=disk:500"
    }

    # GPU processes
    with_label {
      name            = "gpu"
      queue           = "gpu"
      cluster_options = "--gres=gpu:1"
    }
  }

  executor {
    queue_size = 200
  }

  singularity {
    enabled   = true
    cache_dir = "/scratch/singularity"
  }
}

# Example 4: Docker configuration
data "seqera_nextflow_config" "docker" {
  process {
    executor  = "local"
    container = "biocontainers/samtools:1.15"
  }

  docker {
    enabled = true
    remove  = true
  }
}

# Example 5: Advanced with raw_config
data "seqera_nextflow_config" "advanced" {
  process {
    executor = "pbspro"
    queue    = "workq"

    with_label {
      name   = "high_throughput"
      cpus   = 1
      memory = "2 GB"
      time   = "30m"
    }
  }

  executor {
    queue_size = 500
  }

  # Add advanced configuration not in schema
  raw_config = <<-EOF
    // Reporting
    report {
      enabled = true
      file    = 'pipeline_report.html'
    }

    trace {
      enabled = true
      file    = 'pipeline_trace.txt'
      fields  = 'task_id,hash,native_id,name,status,exit,submit,start,complete,duration,realtime,%cpu,rss'
    }

    timeline {
      enabled = true
      file    = 'timeline.html'
    }

    // Manifest
    manifest {
      name        = 'example-pipeline'
      description = 'Example genomics pipeline'
      version     = '1.0.0'
    }
  EOF
}

# Example 6: Composable base + environment-specific
data "seqera_nextflow_config" "base" {
  executor {
    queue_size        = 100
    poll_interval     = "30 sec"
    submit_rate_limit = "10 sec"
  }

  singularity {
    enabled     = true
    auto_mounts = true
    cache_dir   = "/shared/singularity/cache"
  }

  raw_config = <<-EOF
    report.enabled = true
    trace.enabled = true
  EOF
}

data "seqera_nextflow_config" "production" {
  process {
    executor       = "pbspro"
    queue          = "production"
    error_strategy = "retry"
    max_retries    = 3

    with_label {
      name   = "critical"
      queue  = "priority"
    }
  }

  # Append base configuration
  raw_config = data.seqera_nextflow_config.base.config
}

# Example outputs
output "pbspro_config" {
  value       = data.seqera_nextflow_config.pbspro.config
  description = "Generated PBS Pro configuration"
}

output "slurm_config" {
  value       = data.seqera_nextflow_config.slurm.config
  description = "Generated Slurm configuration"
}

output "production_config" {
  value       = data.seqera_nextflow_config.production.config
  description = "Generated production configuration with base settings"
}
