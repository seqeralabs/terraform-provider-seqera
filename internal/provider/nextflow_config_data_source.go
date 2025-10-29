package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/seqeralabs/terraform-provider-seqera/internal/provider/nextflow"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &NextflowConfigDataSource{}

// NewNextflowConfigDataSource creates a new instance of the data source
func NewNextflowConfigDataSource() datasource.DataSource {
	return &NextflowConfigDataSource{}
}

// NextflowConfigDataSource defines the data source implementation
type NextflowConfigDataSource struct{}

// NextflowConfigDataSourceModel describes the data source data model
type NextflowConfigDataSourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Config      types.String               `tfsdk:"config"`
	Process     *nextflow.ProcessModel     `tfsdk:"process"`
	Executor    *nextflow.ExecutorModel    `tfsdk:"executor"`
	Singularity *nextflow.SingularityModel `tfsdk:"singularity"`
	Docker      *nextflow.DockerModel      `tfsdk:"docker"`
	RawConfig   types.String               `tfsdk:"raw_config"`
}

// Metadata returns the data source type name
func (d *NextflowConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nextflow_config"
}

// Schema defines the schema for the data source
func (d *NextflowConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates Nextflow configuration in Groovy format from structured HCL blocks. " +
			"This data source provides a type-safe way to build Nextflow configurations that can be validated at plan time, " +
			"similar to how `aws_iam_policy_document` generates JSON policy documents.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier (hash of the generated configuration)",
			},
			"config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The generated Nextflow configuration in Groovy format",
			},
			"raw_config": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Raw Groovy configuration to append to the generated config. Use this for advanced options not supported by the structured schema.",
			},
		},

		Blocks: map[string]schema.Block{
			"process": schema.SingleNestedBlock{
				MarkdownDescription: "Process configuration block defining default process directives and per-label/per-name overrides",
				Attributes: map[string]schema.Attribute{
					"executor": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Executor type (e.g., 'pbspro', 'slurm', 'awsbatch', 'k8s', 'local')",
					},
					"queue": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Default queue name for job submission",
					},
					"cpus": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Default number of CPUs per process",
					},
					"memory": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Default memory allocation (e.g., '16 GB', '4096 MB')",
					},
					"time": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Default time limit (e.g., '4h', '30m', '1d')",
					},
					"disk": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Default disk space allocation",
					},
					"error_strategy": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Error handling strategy: 'retry', 'ignore', 'terminate', or 'finish'",
					},
					"max_retries": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Maximum number of retry attempts",
					},
					"max_errors": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Maximum number of errors before stopping the pipeline",
					},
					"cluster_options": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Default cluster-specific submission options",
					},
					"container": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Default container image",
					},
					"container_options": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Default container runtime options",
					},
				},
				Blocks: map[string]schema.Block{
					"with_label": schema.SetNestedBlock{
						MarkdownDescription: "Process label-specific overrides using withLabel selector",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Label name (e.g., 'big_mem', 'gpu')",
								},
								"cpus": schema.Int64Attribute{
									Optional:            true,
									MarkdownDescription: "CPU override for this label",
								},
								"memory": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Memory override for this label",
								},
								"time": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Time limit override for this label",
								},
								"disk": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Disk space override for this label",
								},
								"queue": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Queue override for this label",
								},
								"cluster_options": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Cluster options override for this label",
								},
								"container": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Container image override for this label",
								},
								"container_options": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Container options override for this label",
								},
								"error_strategy": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Error strategy override for this label",
								},
								"max_retries": schema.Int64Attribute{
									Optional:            true,
									MarkdownDescription: "Max retries override for this label",
								},
							},
						},
					},
					"with_name": schema.SetNestedBlock{
						MarkdownDescription: "Process name pattern-specific overrides using withName selector",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"pattern": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Process name pattern (e.g., 'align_*', 'STAR', 'GATK_*')",
								},
								"cpus": schema.Int64Attribute{
									Optional:            true,
									MarkdownDescription: "CPU override for processes matching this pattern",
								},
								"memory": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Memory override for processes matching this pattern",
								},
								"time": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Time limit override for processes matching this pattern",
								},
								"disk": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Disk space override for processes matching this pattern",
								},
								"queue": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Queue override for processes matching this pattern",
								},
								"cluster_options": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Cluster options override for processes matching this pattern",
								},
								"container": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Container image override for processes matching this pattern",
								},
								"container_options": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Container options override for processes matching this pattern",
								},
								"error_strategy": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Error strategy override for processes matching this pattern",
								},
								"max_retries": schema.Int64Attribute{
									Optional:            true,
									MarkdownDescription: "Max retries override for processes matching this pattern",
								},
							},
						},
					},
				},
			},

			"executor": schema.SingleNestedBlock{
				MarkdownDescription: "Executor configuration block defining job scheduling and polling behavior",
				Attributes: map[string]schema.Attribute{
					"queue_size": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Maximum number of jobs the executor can submit to the queue simultaneously",
					},
					"poll_interval": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Interval for checking job status (e.g., '30 sec', '1 min')",
					},
					"queue_stat_interval": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Interval for checking queue statistics (e.g., '5 min')",
					},
					"submit_rate_limit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Rate limit for job submission (e.g., '10 sec', '100/1min')",
					},
					"exit_read_timeout": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Timeout for reading exit status from completed jobs",
					},
				},
			},

			"singularity": schema.SingleNestedBlock{
				MarkdownDescription: "Singularity/Apptainer container engine configuration",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Enable Singularity container engine",
					},
					"auto_mounts": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Automatically mount host paths into containers",
					},
					"cache_dir": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Directory for caching container images",
					},
					"run_options": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Additional options for singularity run (e.g., '--bind /data:/data')",
					},
					"env_whitelist": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Comma-separated list of environment variables to pass to containers",
					},
					"pull_timeout": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Timeout for pulling container images (e.g., '20 min')",
					},
				},
			},

			"docker": schema.SingleNestedBlock{
				MarkdownDescription: "Docker container engine configuration",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Enable Docker container engine",
					},
					"registry": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Docker registry URL",
					},
					"run_options": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Additional options for docker run",
					},
					"temp": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Temporary directory for Docker",
					},
					"remove": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Remove containers after execution",
					},
				},
			},
		},
	}
}

// Read executes the data source logic (generates the configuration)
func (d *NextflowConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NextflowConfigDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the Nextflow configuration
	configModel := nextflow.ConfigModel{
		Process:     data.Process,
		Executor:    data.Executor,
		Singularity: data.Singularity,
		Docker:      data.Docker,
		RawConfig:   data.RawConfig,
	}

	generatedConfig := nextflow.BuildConfig(configModel)

	// Set the generated config
	data.Config = types.StringValue(generatedConfig)

	// Generate a hash-based ID
	data.ID = types.StringValue(nextflow.GenerateConfigHash(generatedConfig))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
