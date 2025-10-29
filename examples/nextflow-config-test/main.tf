terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
  }
}

# Provider configuration (not needed for nextflow_config data source, but required by provider schema)
provider "seqera" {
  bearer_auth = "dummy-token-for-testing"
}

# Test the nextflow_config data source
data "seqera_nextflow_config" "test" {
  process {
    executor       = "pbspro"
    queue          = "workq"
    cpus           = 4
    memory         = "16 GB"
    time           = "4h"
    error_strategy = "retry"
    max_retries    = 2

    with_label {
      name            = "big_mem"
      memory          = "128 GB"
      time            = "24h"
      cluster_options = "-l select=1:ncpus=8:mem=128gb"
    }

    with_label {
      name  = "gpu"
      queue = "gpu"
      cluster_options = "-l select=1:ncpus=4:ngpus=1"
    }
  }

  executor {
    queue_size    = 100
    poll_interval = "30 sec"
  }

  singularity {
    enabled     = true
    auto_mounts = true
    cache_dir   = "/shared/singularity/cache"
  }
}

output "generated_config" {
  value = data.seqera_nextflow_config.test.config
}

output "config_id" {
  value = data.seqera_nextflow_config.test.id
}
