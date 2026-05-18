// Package custom_role_data provides the seqera_custom_role data source.
package custom_role_data

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
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
	OrgID        types.Int64    `tfsdk:"org_id"`
	Name         types.String   `tfsdk:"name"`
	Description  types.String   `tfsdk:"description"`
	Permissions  []types.String `tfsdk:"permissions"`
	IsPredefined types.Bool     `tfsdk:"is_predefined"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_role"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a role (predefined or custom) by name. Useful for resolving a " +
			"role name to its permission set without managing the role via Terraform.",
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:    true,
				Description: `Organization numeric identifier.`,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: `Name of the role to look up. Accepts predefined roles (owner, admin, maintain, launch, view, connect) or any custom role name in the organization.`,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `Human-readable description of the role.`,
			},
			"permissions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				MarkdownDescription: "Permission strings the role grants. Predefined-role permission " +
					"sets are stable; custom-role permissions reflect the current platform state.",
			},
			"is_predefined": schema.BoolAttribute{
				Computed:    true,
				Description: `True for built-in roles (owner, admin, maintain, launch, view, connect); false for custom roles.`,
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

	name := data.Name.ValueString()
	res, err := d.client.Roles.DescribeRole(ctx, operations.DescribeRoleRequest{
		RoleName: name,
		OrgID:    data.OrgID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe role", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.Diagnostics.AddError("Role Not Found", fmt.Sprintf("No role found in org %d with name %q.", data.OrgID.ValueInt64(), name))
		return
	}
	if res.StatusCode != 200 {
		common.AddUnexpectedStatus(&resp.Diagnostics, "describing role", res.RawResponse)
		return
	}
	if res.DescribeRoleResponse == nil || res.DescribeRoleResponse.Role == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	role := res.DescribeRoleResponse.Role
	data.Description = types.StringValue(role.Description)
	data.IsPredefined = types.BoolPointerValue(role.IsPredefined)
	data.Permissions = make([]types.String, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		data.Permissions = append(data.Permissions, types.StringValue(p))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
