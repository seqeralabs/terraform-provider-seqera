// Package pipeline_data provides the seqera_pipeline data source.
package pipeline_data

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
	WorkspaceID   types.Int64  `tfsdk:"workspace_id"`
	Name          types.String `tfsdk:"name"`
	PipelineID    types.Int64  `tfsdk:"pipeline_id"`
	Description   types.String `tfsdk:"description"`
	Icon          types.String `tfsdk:"icon"`
	Repository    types.String `tfsdk:"repository"`
	UserID        types.Int64  `tfsdk:"user_id"`
	UserName      types.String `tfsdk:"user_name"`
	UserFirstName types.String `tfsdk:"user_first_name"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up a pipeline by name.`,
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.Int64Attribute{
				Required:    true,
				Description: `Workspace numeric identifier.`,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the pipeline to look up.`,
			},
			"pipeline_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Pipeline numeric identifier.`,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `Description of the pipeline.`,
			},
			"icon": schema.StringAttribute{
				Computed:    true,
				Description: `Icon identifier or URL for the pipeline.`,
			},
			"repository": schema.StringAttribute{
				Computed:    true,
				Description: `Git repository URL containing the pipeline source code.`,
			},
			"user_id": schema.Int64Attribute{
				Computed:    true,
				Description: `User numeric identifier who created the pipeline.`,
			},
			"user_name": schema.StringAttribute{
				Computed:    true,
				Description: `Username of the pipeline creator.`,
			},
			"user_first_name": schema.StringAttribute{
				Computed:    true,
				Description: `First name of the pipeline creator.`,
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
	listRes, err := d.client.Pipelines.ListPipelines(ctx, operations.ListPipelinesRequest{
		WorkspaceID: &workspaceID,
		Search:      &name,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list pipelines", err.Error())
		return
	}
	if listRes.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", listRes.StatusCode))
		return
	}
	if listRes.ListPipelinesResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	// Find the pipeline by exact name match
	// Note: The API does not support pagination. Large workspaces may not return all pipelines.
	var found bool
	for _, p := range listRes.ListPipelinesResponse.Pipelines {
		if p.Name != nil && *p.Name == name {
			data.PipelineID = types.Int64PointerValue(p.PipelineID)
			data.Description = types.StringPointerValue(p.Description)
			data.Icon = types.StringPointerValue(p.Icon)
			data.Repository = types.StringPointerValue(p.Repository)
			data.UserID = types.Int64PointerValue(p.UserID)
			data.UserName = types.StringPointerValue(p.UserName)
			data.UserFirstName = types.StringPointerValue(p.UserFirstName)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Pipeline Not Found", fmt.Sprintf("No pipeline found with name: %s", name))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
