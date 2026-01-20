// Package workspace_participant_data provides the seqera_workspace_participant data source.
package workspace_participant_data

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
	OrgID         types.Int64  `tfsdk:"org_id"`
	WorkspaceID   types.Int64  `tfsdk:"workspace_id"`
	Email         types.String `tfsdk:"email"`
	ParticipantID types.Int64  `tfsdk:"participant_id"`
	MemberID      types.Int64  `tfsdk:"member_id"`
	UserName      types.String `tfsdk:"user_name"`
	FirstName     types.String `tfsdk:"first_name"`
	LastName      types.String `tfsdk:"last_name"`
	Role          types.String `tfsdk:"role"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_participant"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up a workspace participant by email.`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:    true,
				Description: `Organization numeric identifier.`,
			},
			"workspace_id": schema.Int64Attribute{
				Required:    true,
				Description: `Workspace numeric identifier.`,
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: `Email address of the participant to look up.`,
			},
			"participant_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Participant numeric identifier.`,
			},
			"member_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Organization member numeric identifier.`,
			},
			"user_name": schema.StringAttribute{
				Computed:    true,
				Description: `Username of the participant.`,
			},
			"first_name": schema.StringAttribute{
				Computed:    true,
				Description: `First name of the participant.`,
			},
			"last_name": schema.StringAttribute{
				Computed:    true,
				Description: `Last name of the participant.`,
			},
			"role": schema.StringAttribute{
				Computed:    true,
				Description: `Role of the participant in the workspace.`,
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

	email := data.Email.ValueString()
	listRes, err := d.client.Workspaces.ListWorkspaceParticipants(ctx, operations.ListWorkspaceParticipantsRequest{
		OrgID:       data.OrgID.ValueInt64(),
		WorkspaceID: data.WorkspaceID.ValueInt64(),
		Search:      &email,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list workspace participants", err.Error())
		return
	}
	if listRes.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", listRes.StatusCode))
		return
	}
	if listRes.ListParticipantsResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	// Find the participant by exact email match
	// Note: The API does not support pagination. Large workspaces may not return all participants.
	var found bool
	for _, p := range listRes.ListParticipantsResponse.Participants {
		if p.Email != nil && *p.Email == email {
			data.ParticipantID = types.Int64PointerValue(p.ParticipantID)
			data.MemberID = types.Int64PointerValue(p.MemberID)
			data.UserName = types.StringPointerValue(p.UserName)
			data.FirstName = types.StringPointerValue(p.FirstName)
			data.LastName = types.StringPointerValue(p.LastName)
			data.Role = types.StringPointerValue(p.WspRole)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Participant Not Found", fmt.Sprintf("No workspace participant found with email: %s", email))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
