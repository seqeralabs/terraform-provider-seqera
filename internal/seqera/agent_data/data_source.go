// Package agent_data provides the seqera_agent data source.
package agent_data

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/seqeralabs/terraform-provider-seqera/internal/provider/typeconvert"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
	"github.com/seqeralabs/terraform-provider-seqera/internal/seqera/common"
)

var _ datasource.DataSource = &DataSource{}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *sdk.Seqera
}

type DataSourceModel struct {
	WorkspaceID           types.Int64  `tfsdk:"workspace_id"`
	Name                  types.String `tfsdk:"name"`
	ID                    types.String `tfsdk:"id"`
	Status                types.String `tfsdk:"status"`
	Description           types.String `tfsdk:"description"`
	AgentInstructions     types.String `tfsdk:"agent_instructions"`
	ServiceAccountID      types.Int64  `tfsdk:"service_account_id"`
	ServiceAccountName    types.String `tfsdk:"service_account_name"`
	GithubAppCredentialID types.String `tfsdk:"github_app_credential_id"`
	DateCreated           types.String `tfsdk:"date_created"`
	LastUpdated           types.String `tfsdk:"last_updated"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up a workspace agent by name.`,
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.Int64Attribute{
				Required:    true,
				Description: `Workspace numeric identifier.`,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the agent to look up (exact match).`,
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Agent identifier. Use it as `agent.agent_config_id` on `seqera_action`.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status, `active` or `inactive`.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `Short description of the agent.`,
			},
			"agent_instructions": schema.StringAttribute{
				Computed:    true,
				Description: `Instructions the agent follows when it runs.`,
			},
			"service_account_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Service account the agent runs as. Null when none is bound, in which case the agent runs with the identity of whoever triggers it.",
			},
			"service_account_name": schema.StringAttribute{
				Computed:    true,
				Description: `Name of the bound service account, if any.`,
			},
			"github_app_credential_id": schema.StringAttribute{
				Computed:    true,
				Description: `GitHub App credential the agent uses, if any.`,
			},
			"date_created": schema.StringAttribute{
				Computed:    true,
				Description: `Creation timestamp.`,
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: `Last update timestamp.`,
			},
		},
	}
}

func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := common.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if client != nil {
		d.client = client
	}
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := data.WorkspaceID.ValueInt64()
	name := data.Name.ValueString()
	found, err := find(ctx, d.client, workspaceID, name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list agents", err.Error())
		return
	}
	if found == nil {
		resp.Diagnostics.AddError("Agent Not Found", fmt.Sprintf("No agent named %q in workspace %d.", name, workspaceID))
		return
	}

	data.ID = types.StringPointerValue(found.ID)
	data.Status = types.StringNull()
	if found.Status != nil {
		data.Status = types.StringValue(string(*found.Status))
	}
	data.Description = types.StringPointerValue(found.Description)
	data.AgentInstructions = types.StringPointerValue(found.AgentInstructions)
	data.ServiceAccountID = types.Int64PointerValue(found.ServiceAccountID)
	data.ServiceAccountName = types.StringPointerValue(found.ServiceAccountName)
	data.GithubAppCredentialID = types.StringPointerValue(found.GithubAppCredentialID)
	data.DateCreated = types.StringPointerValue(typeconvert.TimePointerToStringPointer(found.DateCreated))
	data.LastUpdated = types.StringPointerValue(typeconvert.TimePointerToStringPointer(found.LastUpdated))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// find pages through the workspace's agents and returns the one whose name
// matches exactly, or nil when none does. The API's search is a substring
// match ("ci" also returns "ci-nightly"), so the exact comparison is required.
// Deleted agents are excluded by the endpoint.
func find(ctx context.Context, client *sdk.Seqera, workspaceID int64, name string) (*shared.AgentDbDto, error) {
	return common.PaginatedSearch(ctx,
		func(ctx context.Context, max, offset int) ([]shared.AgentDbDto, int64, error) {
			res, err := client.Agents.ListAgents(ctx, operations.ListAgentsRequest{
				WorkspaceID: &workspaceID,
				Search:      &name,
				Max:         &max,
				Offset:      &offset,
			})
			if err != nil {
				return nil, 0, err
			}
			if res.StatusCode != 200 {
				return nil, 0, common.UnexpectedStatusErr("listing agents", res.RawResponse)
			}
			if res.ListAgentsResponse == nil {
				return nil, 0, fmt.Errorf("empty response from API")
			}
			var total int64
			if res.ListAgentsResponse.TotalSize != nil {
				total = *res.ListAgentsResponse.TotalSize
			}
			return res.ListAgentsResponse.Agents, total, nil
		},
		func(a *shared.AgentDbDto) bool {
			return a.Name != nil && *a.Name == name
		},
	)
}
