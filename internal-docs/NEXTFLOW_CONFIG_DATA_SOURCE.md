# Nextflow Config Data Source Design

## Overview

The `seqera_nextflow_config` data source generates Nextflow configuration in Groovy format from structured HCL blocks, similar to how `aws_iam_policy_document` generates JSON policy documents from HCL.

This approach provides:

- **Type Safety**: Validate configuration at Terraform plan time instead of runtime
- **Better Integration**: Reference other Terraform resources and use variables
- **Maintainability**: Structured HCL is easier to read and maintain than heredoc Groovy strings
- **Composability**: Build configurations from multiple sources

## Design Philosophy

### Client-Side Only

This is a **pure client-side data source** that does not make any API calls. It only transforms HCL input into Groovy configuration strings. This is similar to:

- `aws_iam_policy_document` (generates JSON)
- `helm_release` (generates YAML values)
- `cloudinit_config` (generates cloud-init config)

### Escape Hatch

Include a `raw_config` attribute to allow users to add unsupported or advanced Groovy configuration that isn't covered by the structured schema.

### Phased Implementation

Start with MVP covering 80% of common use cases, expand based on user feedback.

## MVP Schema (Phase 1)

### Data Source Definition

```hcl
data "seqera_nextflow_config" "example" {
  # Process configuration
  process {
    executor       = "pbspro"
    queue          = "workq"
    cpus           = 4
    memory         = "16 GB"
    time           = "4h"
    error_strategy = "retry"
    max_retries    = 2

    # Label-specific overrides
    with_label {
      name           = "big_mem"
      memory         = "128 GB"
      time           = "24h"
      queue          = "highmem"
      cluster_options = "-l select=1:ncpus=8:mem=128gb"
    }

    with_label {
      name  = "gpu"
      queue = "gpu"
      cluster_options = "-l select=1:ncpus=4:ngpus=1"
    }

    # Process name-specific overrides
    with_name {
      pattern         = "align_*"
      cpus            = 16
      memory          = "64 GB"
      cluster_options = "-l select=1:ncpus=16:mem=64gb"
    }
  }

  # Executor configuration
  executor {
    queue_size           = 100
    poll_interval        = "30 sec"
    queue_stat_interval  = "5 min"
    submit_rate_limit    = "10 sec"
  }

  # Container engine - Singularity
  singularity {
    enabled     = true
    auto_mounts = true
    cache_dir   = "/shared/singularity/cache"
    run_options = "--bind /lustre:/lustre"
  }

  # Container engine - Docker (alternative to Singularity)
  docker {
    enabled  = true
    registry = "docker.io"
  }

  # Escape hatch for unsupported options
  raw_config = <<-EOF
    // Custom Groovy configuration
    timeline.overwrite = true
    dag.overwrite = true
  EOF
}

# Output: Generated configuration
output "nextflow_config" {
  value = data.seqera_nextflow_config.example.config
}
```

### Generated Output

The data source generates valid Groovy configuration:

```groovy
process {
  executor = 'pbspro'
  queue = 'workq'
  cpus = 4
  memory = '16 GB'
  time = '4h'
  errorStrategy = 'retry'
  maxRetries = 2

  withLabel: big_mem {
    memory = '128 GB'
    time = '24h'
    queue = 'highmem'
    clusterOptions = '-l select=1:ncpus=8:mem=128gb'
  }

  withLabel: gpu {
    queue = 'gpu'
    clusterOptions = '-l select=1:ncpus=4:ngpus=1'
  }

  withName: 'align_*' {
    cpus = 16
    memory = '64 GB'
    clusterOptions = '-l select=1:ncpus=16:mem=64gb'
  }
}

executor {
  queueSize = 100
  pollInterval = '30 sec'
  queueStatInterval = '5 min'
  submitRateLimit = '10 sec'
}

singularity {
  enabled = true
  autoMounts = true
  cacheDir = '/shared/singularity/cache'
  runOptions = '--bind /lustre:/lustre'
}

// Custom Groovy configuration
timeline.overwrite = true
dag.overwrite = true
```

## Schema Definition

### Root Attributes

| Attribute    | Type   | Required | Description                        |
| ------------ | ------ | -------- | ---------------------------------- |
| `id`         | String | Computed | Unique identifier (hash of config) |
| `config`     | String | Computed | Generated Groovy configuration     |
| `raw_config` | String | Optional | Raw Groovy code appended to config |

### Process Block

| Attribute           | Type   | Optional | Description                                        |
| ------------------- | ------ | -------- | -------------------------------------------------- |
| `executor`          | String | Yes      | Executor type (pbspro, slurm, awsbatch, k8s, etc.) |
| `queue`             | String | Yes      | Default queue name                                 |
| `cpus`              | Number | Yes      | Default CPU allocation                             |
| `memory`            | String | Yes      | Default memory allocation (e.g., "16 GB")          |
| `time`              | String | Yes      | Default time limit (e.g., "4h")                    |
| `disk`              | String | Yes      | Default disk space                                 |
| `error_strategy`    | String | Yes      | Error strategy: retry, ignore, terminate, finish   |
| `max_retries`       | Number | Yes      | Maximum number of retries                          |
| `max_errors`        | Number | Yes      | Maximum number of errors before stopping           |
| `cluster_options`   | String | Yes      | Default cluster-specific options                   |
| `container`         | String | Yes      | Default container image                            |
| `container_options` | String | Yes      | Default container options                          |

### Process.with_label Block (repeatable)

| Attribute           | Type   | Required | Description                         |
| ------------------- | ------ | -------- | ----------------------------------- |
| `name`              | String | Yes      | Label name (e.g., "big_mem", "gpu") |
| `cpus`              | Number | Optional | CPU override                        |
| `memory`            | String | Optional | Memory override                     |
| `time`              | String | Optional | Time override                       |
| `disk`              | String | Optional | Disk override                       |
| `queue`             | String | Optional | Queue override                      |
| `cluster_options`   | String | Optional | Cluster options override            |
| `container`         | String | Optional | Container override                  |
| `container_options` | String | Optional | Container options override          |
| `error_strategy`    | String | Optional | Error strategy override             |
| `max_retries`       | Number | Optional | Max retries override                |

### Process.with_name Block (repeatable)

Same attributes as `with_label` but:

- `pattern` (String, Required): Process name pattern (e.g., "align\_\*", "STAR")

### Executor Block

| Attribute             | Type   | Optional | Description                     |
| --------------------- | ------ | -------- | ------------------------------- |
| `queue_size`          | Number | Yes      | Max concurrent jobs             |
| `poll_interval`       | String | Yes      | Job status check interval       |
| `queue_stat_interval` | String | Yes      | Queue statistics check interval |
| `submit_rate_limit`   | String | Yes      | Job submission rate limit       |
| `exit_read_timeout`   | String | Yes      | Timeout for reading exit status |

### Singularity Block

| Attribute       | Type    | Optional | Description                             |
| --------------- | ------- | -------- | --------------------------------------- |
| `enabled`       | Boolean | Yes      | Enable Singularity                      |
| `auto_mounts`   | Boolean | Yes      | Auto-mount host paths                   |
| `cache_dir`     | String  | Yes      | Container cache directory               |
| `run_options`   | String  | Yes      | Additional run options (e.g., "--bind") |
| `env_whitelist` | String  | Yes      | Environment variables to pass           |
| `pull_timeout`  | String  | Yes      | Timeout for pulling images              |

### Docker Block

| Attribute     | Type    | Optional | Description                       |
| ------------- | ------- | -------- | --------------------------------- |
| `enabled`     | Boolean | Yes      | Enable Docker                     |
| `registry`    | String  | Yes      | Docker registry URL               |
| `run_options` | String  | Yes      | Additional run options            |
| `temp`        | String  | Yes      | Temporary directory               |
| `remove`      | Boolean | Yes      | Remove containers after execution |

## HCL to Groovy Mapping Rules

### Field Name Conversion

HCL snake_case → Groovy camelCase:

- `error_strategy` → `errorStrategy`
- `max_retries` → `maxRetries`
- `queue_size` → `queueSize`
- `cluster_options` → `clusterOptions`

### String Values

All string values are single-quoted in Groovy:

- HCL: `executor = "pbspro"` → Groovy: `executor = 'pbspro'`

### Boolean Values

HCL booleans map directly:

- HCL: `enabled = true` → Groovy: `enabled = true`

### Number Values

Numbers map directly:

- HCL: `cpus = 4` → Groovy: `cpus = 4`

### Block Syntax

HCL blocks map to Groovy closures:

```hcl
process {
  cpus = 4
}
```

→

```groovy
process {
  cpus = 4
}
```

### Label/Name Directives

HCL nested blocks map to Groovy label syntax:

```hcl
with_label {
  name = "big_mem"
  memory = "128 GB"
}
```

→

```groovy
withLabel: big_mem {
  memory = '128 GB'
}
```

### Pattern Matching

For `with_name`, patterns need special handling:

```hcl
with_name {
  pattern = "align_*"
  cpus = 16
}
```

→

```groovy
withName: 'align_*' {
  cpus = 16
}
```

## Implementation Structure

### File Organization

```
internal/provider/
├── nextflow_config_data_source.go       # Data source implementation
├── nextflow_config_data_source_test.go  # Tests
└── nextflow/                             # Config builder package
    ├── builder.go                        # Main builder
    ├── process.go                        # Process block builder
    ├── executor.go                       # Executor block builder
    ├── container.go                      # Container engine builders
    └── types.go                          # Shared types and constants
```

### Key Functions

```go
// builder.go
func BuildConfig(model NextflowConfigModel) (string, error)

// process.go
func buildProcessBlock(process ProcessModel) string
func buildWithLabel(label WithLabelModel) string
func buildWithName(name WithNameModel) string

// executor.go
func buildExecutorBlock(executor ExecutorModel) string

// container.go
func buildSingularityBlock(singularity SingularityModel) string
func buildDockerBlock(docker DockerModel) string

// types.go
func snakeToCamel(s string) string
func escapeGroovyString(s string) string
func generateConfigHash(config string) string
```

## Usage Examples

### Example 1: HPC with PBS Pro

```hcl
data "seqera_nextflow_config" "pbspro" {
  process {
    executor = "pbspro"
    queue    = "workq"
    cpus     = 4
    memory   = "16 GB"
    time     = "4h"

    with_label {
      name            = "big_mem"
      memory          = "128 GB"
      time            = "24h"
      cluster_options = "-l select=1:ncpus=8:mem=128gb"
    }
  }

  executor {
    queue_size = 100
  }

  singularity {
    enabled   = true
    cache_dir = "/shared/singularity/cache"
  }
}

resource "seqera_compute_altair_pbs_pro" "main" {
  name             = "pbspro-compute"
  workspace_id     = seqera_workspace.main.id
  credentials_id   = seqera_altair_pbs_pro_credential.main.credentials_id
  work_directory   = "/shared/work"

  # Use the generated config
  nextflow_config = data.seqera_nextflow_config.pbspro.config
}
```

### Example 2: Composable Configs

```hcl
# Base configuration
data "seqera_nextflow_config" "base" {
  executor {
    queue_size = 100
    poll_interval = "30 sec"
  }

  singularity {
    enabled   = true
    cache_dir = "/shared/singularity/cache"
  }
}

# Environment-specific configuration
data "seqera_nextflow_config" "production" {
  process {
    executor = "pbspro"
    queue    = "production"

    error_strategy = "retry"
    max_retries    = 3
  }

  # Append base config
  raw_config = data.seqera_nextflow_config.base.config
}
```

### Example 3: Multiple Compute Types

```hcl
# Slurm configuration
data "seqera_nextflow_config" "slurm" {
  process {
    executor = "slurm"
    queue    = "compute"

    with_label {
      name  = "gpu"
      queue = "gpu"
      cluster_options = "--gres=gpu:1"
    }
  }
}

# AWS Batch configuration
data "seqera_nextflow_config" "aws" {
  process {
    executor = "awsbatch"
    queue    = "nextflow-queue"
  }
}
```

## Testing Strategy

### Unit Tests

Test config generation with various inputs:

```go
func TestProcessBlockGeneration(t *testing.T)
func TestWithLabelGeneration(t *testing.T)
func TestExecutorBlockGeneration(t *testing.T)
func TestSingularityBlockGeneration(t *testing.T)
func TestRawConfigAppending(t *testing.T)
func TestSpecialCharacterEscaping(t *testing.T)
```

### Integration Tests

Test with real compute environments in `examples/tests/`:

```terraform
terraform {
  required_providers {
    seqera = {
      source = "terraform.local/local/seqera"
    }
  }
}

data "seqera_nextflow_config" "test" {
  process {
    executor = "local"
    cpus     = 2
  }
}

output "config" {
  value = data.seqera_nextflow_config.test.config
}
```

### Validation Tests

Validate generated Groovy syntax:

1. Parse with Groovy parser (if available)
2. Test with actual Nextflow execution
3. Compare output with known-good configs

## Future Enhancements (Post-MVP)

### Cloud Provider Blocks

```hcl
aws {
  region = "us-east-1"
  batch {
    max_parallel_transfers = 4
    volumes = ["/host/path:/container/path"]
  }
}

google {
  project = "my-project"
  region  = "us-central1"
  batch {
    spot = true
  }
}

azure {
  batch {
    auto_storage = true
  }
}
```

### Kubernetes Block

```hcl
k8s {
  namespace       = "nextflow"
  service_account = "nextflow-runner"

  pod {
    node_selector = {
      workload = "nextflow"
    }

    annotation {
      key   = "cluster-autoscaler.kubernetes.io/safe-to-evict"
      value = "false"
    }
  }

  storage_claim_name = "nextflow-pvc"
}
```

### Reporting Blocks

```hcl
report {
  enabled = true
  file    = "report.html"
}

trace {
  enabled = true
  file    = "trace.txt"
  fields  = "task_id,hash,native_id,name,status"
}

timeline {
  enabled = true
  file    = "timeline.html"
}

dag {
  enabled = true
  file    = "dag.html"
}
```

### Manifest Block

```hcl
manifest {
  name                = "my-pipeline"
  description         = "Pipeline description"
  version             = "1.0.0"
  nextflow_version    = ">=22.10.0"
  author              = "Author Name"
  homepage            = "https://github.com/user/pipeline"
  main_script         = "main.nf"
  default_branch      = "main"
}
```

### Profiles

```hcl
profile {
  name = "docker"

  docker {
    enabled = true
  }
}

profile {
  name = "singularity"

  singularity {
    enabled = true
  }
}
```

## Open Questions

1. **Validation**: Should we validate Groovy syntax or rely on Nextflow runtime?
2. **Defaults**: Should we set sensible defaults or require explicit values?
3. **Composition**: Support merging multiple data sources?
4. **Format**: Support JSON output in addition to Groovy?
5. **Advanced Features**: Support custom functions or variables in Groovy?

## Success Metrics

- Users can replace 80% of heredoc config strings with structured HCL
- Generated configs are syntactically valid Groovy
- Data source survives Speakeasy regeneration
- Performance: Config generation takes <100ms
- Maintainability: Adding new fields is straightforward

## References

- [Nextflow Config Reference](https://www.nextflow.io/docs/latest/config.html)
- [Terraform Data Sources](https://developer.hashicorp.com/terraform/plugin/framework/data-sources)
- [aws_iam_policy_document Pattern](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document)
- [Speakeasy Custom Data Sources](https://www.speakeasy.com/docs/terraform/provider-configuration#custom-resources-or-data-sources)
