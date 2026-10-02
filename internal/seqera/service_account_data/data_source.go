// Package service_account_data provides the seqera_service_account data source.
package service_account_data

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
	OrgID       types.Int64  `tfsdk:"org_id"`
	Name        types.String `tfsdk:"name"`
	ID          types.Int64  `tfsdk:"id"`
	MemberID    types.Int64  `tfsdk:"member_id"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up an organization service account by name.`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:    true,
				Description: `Organization numeric identifier.`,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the service account to look up.`,
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Service account numeric identifier. Use it as `seqera_agent.service_account_id`.",
			},
			"member_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Organization membership identifier. Use it as `seqera_workspace_participant.member_id`.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `Description of the service account.`,
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: `Creation timestamp.`,
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

	orgID := data.OrgID.ValueInt64()
	name := data.Name.ValueString()
	found, err := find(ctx, d.client, orgID, name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list service accounts", err.Error())
		return
	}
	if found == nil {
		resp.Diagnostics.AddError("Service Account Not Found", fmt.Sprintf("No service account named %q in org %d.", name, orgID))
		return
	}

	data.ID = types.Int64PointerValue(found.ID)
	data.MemberID = types.Int64PointerValue(found.MemberID)
	data.Description = types.StringPointerValue(found.Description)
	data.CreatedAt = types.StringPointerValue(typeconvert.TimePointerToStringPointer(found.CreatedAt))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// find pages through the organization's service accounts and returns the one
// whose name matches exactly, or nil when none does. The list endpoint has no
// name filter, so matching happens client-side.
func find(ctx context.Context, client *sdk.Seqera, orgID int64, name string) (*shared.ServiceAccountDto, error) {
	return common.PaginatedSearch(ctx,
		func(ctx context.Context, max, offset int) ([]shared.ServiceAccountDto, int64, error) {
			res, err := client.ServiceAccounts.ListServiceAccounts(ctx, operations.ListServiceAccountsRequest{
				OrgID:  orgID,
				Max:    &max,
				Offset: &offset,
			})
			if err != nil {
				return nil, 0, err
			}
			if res.StatusCode != 200 {
				return nil, 0, common.UnexpectedStatusErr("listing service accounts", res.RawResponse)
			}
			if res.ListServiceAccountsResponse == nil {
				return nil, 0, fmt.Errorf("empty response from API")
			}
			var total int64
			if res.ListServiceAccountsResponse.TotalSize != nil {
				total = *res.ListServiceAccountsResponse.TotalSize
			}
			return res.ListServiceAccountsResponse.ServiceAccounts, total, nil
		},
		func(sa *shared.ServiceAccountDto) bool {
			return sa.Name != nil && *sa.Name == name
		},
	)
}
