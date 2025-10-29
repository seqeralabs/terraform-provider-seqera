package nextflow

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProcessModel represents the process configuration block
type ProcessModel struct {
	Executor         types.String      `tfsdk:"executor"`
	Queue            types.String      `tfsdk:"queue"`
	Cpus             types.Int64       `tfsdk:"cpus"`
	Memory           types.String      `tfsdk:"memory"`
	Time             types.String      `tfsdk:"time"`
	Disk             types.String      `tfsdk:"disk"`
	ErrorStrategy    types.String      `tfsdk:"error_strategy"`
	MaxRetries       types.Int64       `tfsdk:"max_retries"`
	MaxErrors        types.Int64       `tfsdk:"max_errors"`
	ClusterOptions   types.String      `tfsdk:"cluster_options"`
	Container        types.String      `tfsdk:"container"`
	ContainerOptions types.String      `tfsdk:"container_options"`
	WithLabel        []WithLabelModel  `tfsdk:"with_label"`
	WithName         []WithNameModel   `tfsdk:"with_name"`
}

// WithLabelModel represents a process label override
type WithLabelModel struct {
	Name             types.String `tfsdk:"name"`
	Cpus             types.Int64  `tfsdk:"cpus"`
	Memory           types.String `tfsdk:"memory"`
	Time             types.String `tfsdk:"time"`
	Disk             types.String `tfsdk:"disk"`
	Queue            types.String `tfsdk:"queue"`
	ClusterOptions   types.String `tfsdk:"cluster_options"`
	Container        types.String `tfsdk:"container"`
	ContainerOptions types.String `tfsdk:"container_options"`
	ErrorStrategy    types.String `tfsdk:"error_strategy"`
	MaxRetries       types.Int64  `tfsdk:"max_retries"`
}

// WithNameModel represents a process name pattern override
type WithNameModel struct {
	Pattern          types.String `tfsdk:"pattern"`
	Cpus             types.Int64  `tfsdk:"cpus"`
	Memory           types.String `tfsdk:"memory"`
	Time             types.String `tfsdk:"time"`
	Disk             types.String `tfsdk:"disk"`
	Queue            types.String `tfsdk:"queue"`
	ClusterOptions   types.String `tfsdk:"cluster_options"`
	Container        types.String `tfsdk:"container"`
	ContainerOptions types.String `tfsdk:"container_options"`
	ErrorStrategy    types.String `tfsdk:"error_strategy"`
	MaxRetries       types.Int64  `tfsdk:"max_retries"`
}

// ExecutorModel represents the executor configuration block
type ExecutorModel struct {
	QueueSize          types.Int64  `tfsdk:"queue_size"`
	PollInterval       types.String `tfsdk:"poll_interval"`
	QueueStatInterval  types.String `tfsdk:"queue_stat_interval"`
	SubmitRateLimit    types.String `tfsdk:"submit_rate_limit"`
	ExitReadTimeout    types.String `tfsdk:"exit_read_timeout"`
}

// SingularityModel represents the Singularity container engine configuration
type SingularityModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	AutoMounts    types.Bool   `tfsdk:"auto_mounts"`
	CacheDir      types.String `tfsdk:"cache_dir"`
	RunOptions    types.String `tfsdk:"run_options"`
	EnvWhitelist  types.String `tfsdk:"env_whitelist"`
	PullTimeout   types.String `tfsdk:"pull_timeout"`
}

// DockerModel represents the Docker container engine configuration
type DockerModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	Registry   types.String `tfsdk:"registry"`
	RunOptions types.String `tfsdk:"run_options"`
	Temp       types.String `tfsdk:"temp"`
	Remove     types.Bool   `tfsdk:"remove"`
}

// ConfigModel represents the complete Nextflow configuration data source model
type ConfigModel struct {
	Process     *ProcessModel     `tfsdk:"process"`
	Executor    *ExecutorModel    `tfsdk:"executor"`
	Singularity *SingularityModel `tfsdk:"singularity"`
	Docker      *DockerModel      `tfsdk:"docker"`
	RawConfig   types.String      `tfsdk:"raw_config"`
}

// BuildConfig generates a Nextflow configuration string from the model
func BuildConfig(model ConfigModel) string {
	var config strings.Builder

	// Process block
	if model.Process != nil {
		config.WriteString(buildProcessBlock(*model.Process))
	}

	// Executor block
	if model.Executor != nil {
		config.WriteString(buildExecutorBlock(*model.Executor))
	}

	// Singularity block
	if model.Singularity != nil {
		config.WriteString(buildSingularityBlock(*model.Singularity))
	}

	// Docker block
	if model.Docker != nil {
		config.WriteString(buildDockerBlock(*model.Docker))
	}

	// Raw config (appended last)
	if !model.RawConfig.IsNull() && !model.RawConfig.IsUnknown() {
		rawConfig := model.RawConfig.ValueString()
		if rawConfig != "" {
			config.WriteString("\n")
			config.WriteString(rawConfig)
			if !strings.HasSuffix(rawConfig, "\n") {
				config.WriteString("\n")
			}
		}
	}

	return config.String()
}
