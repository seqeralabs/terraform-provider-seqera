// Package team_data provides the seqera_team data source.
package team_data

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
	OrgID        types.Int64  `tfsdk:"org_id"`
	Name         types.String `tfsdk:"name"`
	ID           types.Int64  `tfsdk:"id"`
	TeamID       types.Int64  `tfsdk:"team_id"`
	Description  types.String `tfsdk:"description"`
	MembersCount types.Int64  `tfsdk:"members_count"`
	AvatarURL    types.String `tfsdk:"avatar_url"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up an organization team by name.`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:    true,
				Description: `Organization numeric identifier.`,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the team to look up.`,
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Team numeric identifier. Alias of `team_id` — matches the `team_id` argument expected by `seqera_workspace_participant`.",
			},
			"team_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Team numeric identifier.`,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `Description of the team.`,
			},
			"members_count": schema.Int64Attribute{
				Computed:    true,
				Description: `Total number of members in the team.`,
			},
			"avatar_url": schema.StringAttribute{
				Computed:    true,
				Description: `URL to the team's avatar or profile image.`,
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
	listRes, err := d.client.Teams.ListOrganizationTeams(ctx, operations.ListOrganizationTeamsRequest{
		OrgID: data.OrgID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list teams", err.Error())
		return
	}
	if listRes.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", listRes.StatusCode))
		return
	}
	if listRes.ListTeamResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	// Find the team by exact name match.
	// The platform's ListOrganizationTeams endpoint has no name filter, so
	// all teams are fetched and filtered locally.
	var found bool
	for _, t := range listRes.ListTeamResponse.Teams {
		if t.Name != nil && *t.Name == name {
			data.TeamID = types.Int64PointerValue(t.TeamID)
			data.ID = types.Int64PointerValue(t.TeamID)
			data.Description = types.StringPointerValue(t.Description)
			data.MembersCount = types.Int64PointerValue(t.MembersCount)
			data.AvatarURL = types.StringPointerValue(t.AvatarURL)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Team Not Found", fmt.Sprintf("No team found in org %d with name %q.", data.OrgID.ValueInt64(), name))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
