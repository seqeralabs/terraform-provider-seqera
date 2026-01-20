// Package pipeline_secret_data provides the seqera_pipeline_secret data source.
package pipeline_secret_data

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
)

var _ datasource.DataSource = &DataSource{}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *sdk.Seqera
}

type DataSourceModel struct {
	WorkspaceID types.Int64  `tfsdk:"workspace_id"`
	Name        types.String `tfsdk:"name"`
	SecretID    types.Int64  `tfsdk:"secret_id"`
	LastUsed    types.String `tfsdk:"last_used"`
	DateCreated types.String `tfsdk:"date_created"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_secret"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up a pipeline secret by name.`,
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.Int64Attribute{
				Required:    true,
				Description: `Workspace numeric identifier.`,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the pipeline secret to look up.`,
			},
			"secret_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Pipeline secret numeric identifier.`,
			},
			"last_used": schema.StringAttribute{
				Computed:    true,
				Description: `Timestamp when the secret was last accessed by a workflow.`,
			},
			"date_created": schema.StringAttribute{
				Computed:    true,
				Description: `Timestamp when the secret was created.`,
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: `Timestamp when the secret was last updated.`,
			},
		},
	}
}

func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Seqera)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.Seqera, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	workspaceID := data.WorkspaceID.ValueInt64()
	listRes, err := d.client.PipelineSecrets.ListPipelineSecrets(ctx, operations.ListPipelineSecretsRequest{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list pipeline secrets", err.Error())
		return
	}
	if listRes.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", listRes.StatusCode))
		return
	}
	if listRes.ListPipelineSecretsResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	// Find the pipeline secret by exact name match
	// Note: The ListPipelineSecrets API does not support a search filter, so all secrets are fetched and filtered locally.
	// The API also does not support pagination. Large workspaces may not return all pipeline secrets.
	var found bool
	for _, s := range listRes.ListPipelineSecretsResponse.PipelineSecrets {
		if s.Name == name {
			data.SecretID = types.Int64PointerValue(s.ID)
			if s.LastUsed != nil {
				data.LastUsed = types.StringValue(s.LastUsed.Format("2006-01-02T15:04:05Z07:00"))
			} else {
				data.LastUsed = types.StringNull()
			}
			if s.DateCreated != nil {
				data.DateCreated = types.StringValue(s.DateCreated.Format("2006-01-02T15:04:05Z07:00"))
			} else {
				data.DateCreated = types.StringNull()
			}
			if s.LastUpdated != nil {
				data.LastUpdated = types.StringValue(s.LastUpdated.Format("2006-01-02T15:04:05Z07:00"))
			} else {
				data.LastUpdated = types.StringNull()
			}
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Pipeline Secret Not Found", fmt.Sprintf("No pipeline secret found with name: %s", name))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
