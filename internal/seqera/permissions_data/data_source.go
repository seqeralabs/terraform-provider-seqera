// Package permissions_data provides the seqera_permissions data source —
// the catalogue of grants assignable to custom roles.
package permissions_data

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	Category    types.String `tfsdk:"category"`
	Permissions types.List   `tfsdk:"permissions"`
	Names       types.List   `tfsdk:"names"`
	Categories  types.List   `tfsdk:"categories"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Catalogue of permissions that can be assigned to a `seqera_custom_role`. " +
			"Use this to introspect what grants the platform supports, or to validate that a permission " +
			"string is real before applying.",
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:    true,
				Description: `Organization numeric identifier.`,
			},
			"category": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Optional client-side filter. When set, only permissions whose category matches " +
					"(case-insensitive) are returned. Categories observed in the catalogue: `Compute`, `Data`, " +
					"`Pipelines`, `Settings`, `Studios`.",
			},
			"permissions": schema.ListNestedAttribute{
				Computed:    true,
				Description: `Full permission catalogue. Each entry has a name and a category.`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: `Permission name (e.g. "pipeline:read", "workflow:execute").`,
						},
						"category": schema.StringAttribute{
							Computed:    true,
							Description: `Category the permission belongs to.`,
						},
					},
				},
			},
			"names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				MarkdownDescription: "Flat sorted list of just the permission names. Convenient for " +
					"`contains(data.seqera_permissions.all.names, \"pipeline:read\")` validation checks.",
			},
			"categories": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: `Sorted list of distinct categories observed in the catalogue.`,
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

	res, err := d.client.Roles.ListRolePermissions(ctx, operations.ListRolePermissionsRequest{
		OrgID: data.OrgID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list role permissions", err.Error())
		return
	}
	if res.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API response", fmt.Sprintf("Status code: %d", res.StatusCode))
		return
	}
	if res.ListRolePermissionsResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Empty response from API")
		return
	}

	wantCategory := data.Category.ValueString()
	wantFilter := !data.Category.IsNull() && wantCategory != ""

	type pair struct{ name, category string }
	pairs := make([]pair, 0, len(res.ListRolePermissionsResponse.Permissions))
	categorySet := map[string]struct{}{}
	for _, p := range res.ListRolePermissionsResponse.Permissions {
		name := ""
		if p.Name != nil {
			name = *p.Name
		}
		cat := ""
		if p.Category != nil {
			cat = *p.Category
		}
		if wantFilter && !equalFold(cat, wantCategory) {
			continue
		}
		pairs = append(pairs, pair{name, cat})
		if cat != "" {
			categorySet[cat] = struct{}{}
		}
	}

	// Stable ordering so plans don't churn when the API returns
	// permissions in an arbitrary order.
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].name < pairs[j].name })

	permObjs := make([]attr.Value, 0, len(pairs))
	names := make([]attr.Value, 0, len(pairs))
	permObjType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":     types.StringType,
		"category": types.StringType,
	}}
	for _, p := range pairs {
		obj, diags := types.ObjectValue(permObjType.AttrTypes, map[string]attr.Value{
			"name":     types.StringValue(p.name),
			"category": types.StringValue(p.category),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		permObjs = append(permObjs, obj)
		names = append(names, types.StringValue(p.name))
	}

	categoriesList := make([]string, 0, len(categorySet))
	for c := range categorySet {
		categoriesList = append(categoriesList, c)
	}
	sort.Strings(categoriesList)
	catValues := make([]attr.Value, 0, len(categoriesList))
	for _, c := range categoriesList {
		catValues = append(catValues, types.StringValue(c))
	}

	permsList, diags := types.ListValue(permObjType, permObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Permissions = permsList

	namesList, diags := types.ListValue(types.StringType, names)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Names = namesList

	catsList, diags := types.ListValue(types.StringType, catValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Categories = catsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// equalFold avoids importing strings just for one call site.
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
