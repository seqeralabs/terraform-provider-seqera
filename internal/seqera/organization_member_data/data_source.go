// Package organization_member_data provides the seqera_organization_member data source.
package organization_member_data

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
	OrgID     types.Int64  `tfsdk:"org_id"`
	Email     types.String `tfsdk:"email"`
	MemberID  types.Int64  `tfsdk:"member_id"`
	UserID    types.Int64  `tfsdk:"user_id"`
	UserName  types.String `tfsdk:"user_name"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Role      types.String `tfsdk:"role"`
	Avatar    types.String `tfsdk:"avatar"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up an organization member by email.`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:    true,
				Description: `Organization numeric identifier.`,
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: `Email address of the member to look up.`,
			},
			"member_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Organization member numeric identifier.`,
			},
			"user_id": schema.Int64Attribute{
				Computed:    true,
				Description: `User numeric identifier.`,
			},
			"user_name": schema.StringAttribute{
				Computed:    true,
				Description: `Username of the member.`,
			},
			"first_name": schema.StringAttribute{
				Computed:    true,
				Description: `First name of the member.`,
			},
			"last_name": schema.StringAttribute{
				Computed:    true,
				Description: `Last name of the member.`,
			},
			"role": schema.StringAttribute{
				Computed:    true,
				Description: `Role of the member (owner, member, collaborator).`,
			},
			"avatar": schema.StringAttribute{
				Computed:    true,
				Description: `Avatar URL of the member.`,
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
	listRes, err := d.client.Orgs.ListOrganizationMembers(ctx, operations.ListOrganizationMembersRequest{
		OrgID:  data.OrgID.ValueInt64(),
		Search: &email,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list organization members", err.Error())
		return
	}
	if listRes.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", listRes.StatusCode))
		return
	}
	if listRes.ListMembersResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	// Find the member by exact email match
	// Note: The API does not support pagination. Large organizations may not return all members.
	var found bool
	for _, m := range listRes.ListMembersResponse.Members {
		if m.Email != nil && *m.Email == email {
			data.MemberID = types.Int64PointerValue(m.MemberID)
			data.UserID = types.Int64PointerValue(m.UserID)
			data.UserName = types.StringPointerValue(m.UserName)
			data.FirstName = types.StringPointerValue(m.FirstName)
			data.LastName = types.StringPointerValue(m.LastName)
			data.Avatar = types.StringPointerValue(m.Avatar)
			if m.Role != nil {
				data.Role = types.StringValue(string(*m.Role))
			}
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Member Not Found", fmt.Sprintf("No organization member found with email: %s", email))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
