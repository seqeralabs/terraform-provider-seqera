// Package organization_data provides the seqera_organization data source.
package organization_data

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
	Name        types.String `tfsdk:"name"`
	OrgID       types.Int64  `tfsdk:"org_id"`
	FullName    types.String `tfsdk:"full_name"`
	Description types.String `tfsdk:"description"`
	Location    types.String `tfsdk:"location"`
	Website     types.String `tfsdk:"website"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Look up an organization by name.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the organization to look up.`,
			},
			"org_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Organization numeric identifier.`,
			},
			"full_name": schema.StringAttribute{
				Computed:    true,
				Description: `Full name of the organization.`,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `Description of the organization.`,
			},
			"location": schema.StringAttribute{
				Computed:    true,
				Description: `Geographic location or address of the organization.`,
			},
			"website": schema.StringAttribute{
				Computed:    true,
				Description: `Official website URL for the organization.`,
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
	listRes, err := d.client.Orgs.ListOrganizations(ctx, operations.ListOrganizationsRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list organizations", err.Error())
		return
	}
	if listRes.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", listRes.StatusCode))
		return
	}
	if listRes.ListOrganizationsResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	// Find the organization by exact name match
	// Note: The ListOrganizations API does not support a search filter, so all organizations are fetched and filtered locally.
	var found bool
	for _, org := range listRes.ListOrganizationsResponse.Organizations {
		if org.Name != nil && *org.Name == name {
			data.OrgID = types.Int64PointerValue(org.OrgID)
			data.FullName = types.StringPointerValue(org.FullName)
			data.Description = types.StringPointerValue(org.Description)
			data.Location = types.StringPointerValue(org.Location)
			data.Website = types.StringPointerValue(org.Website)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Organization Not Found", fmt.Sprintf("No organization found with name: %s", name))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
