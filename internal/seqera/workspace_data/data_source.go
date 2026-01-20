// Package workspace_data provides the seqera_workspace data source.
package workspace_data

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
	OrgID       types.Int64  `tfsdk:"org_id"`
	Name        types.String `tfsdk:"name"`
	WorkspaceID types.Int64  `tfsdk:"workspace_id"`
	FullName    types.String `tfsdk:"full_name"`
	Description types.String `tfsdk:"description"`
	Visibility  types.String `tfsdk:"visibility"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up a workspace by name.`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:    true,
				Description: `Organization numeric identifier.`,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the workspace to look up.`,
			},
			"workspace_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Workspace numeric identifier.`,
			},
			"full_name": schema.StringAttribute{
				Computed:    true,
				Description: `Full name of the workspace.`,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `Description of the workspace.`,
			},
			"visibility": schema.StringAttribute{
				Computed:    true,
				Description: `Visibility of the workspace (PRIVATE or PUBLIC).`,
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
	listRes, err := d.client.Workspaces.ListWorkspaces(ctx, operations.ListWorkspacesRequest{
		OrgID: data.OrgID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list workspaces", err.Error())
		return
	}
	if listRes.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", listRes.StatusCode))
		return
	}
	if listRes.ListWorkspacesResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	// Find the workspace by exact name match
	// Note: The ListWorkspaces API does not support a search filter, so all workspaces are fetched and filtered locally.
	// The API also does not support pagination. Large organizations may not return all workspaces.
	var found bool
	for _, w := range listRes.ListWorkspacesResponse.Workspaces {
		if w.Name != nil && *w.Name == name {
			data.WorkspaceID = types.Int64PointerValue(w.ID)
			data.FullName = types.StringPointerValue(w.FullName)
			data.Description = types.StringPointerValue(w.Description)
			if w.Visibility != nil {
				data.Visibility = types.StringValue(string(*w.Visibility))
			}
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Workspace Not Found", fmt.Sprintf("No workspace found with name: %s", name))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
