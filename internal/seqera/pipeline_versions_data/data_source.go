// Package pipeline_versions_data provides the seqera_pipeline_versions data source.
package pipeline_versions_data

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

type versionModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	IsDefault        types.Bool   `tfsdk:"is_default"`
	Hash             types.String `tfsdk:"hash"`
	DateCreated      types.String `tfsdk:"date_created"`
	LastUpdated      types.String `tfsdk:"last_updated"`
	CreatorUserID    types.Int64  `tfsdk:"creator_user_id"`
	CreatorUserName  types.String `tfsdk:"creator_user_name"`
	CreatorFirstName types.String `tfsdk:"creator_first_name"`
	CreatorLastName  types.String `tfsdk:"creator_last_name"`
	CreatorAvatarURL types.String `tfsdk:"creator_avatar_url"`
}

type DataSourceModel struct {
	PipelineID  types.Int64    `tfsdk:"pipeline_id"`
	WorkspaceID types.Int64    `tfsdk:"workspace_id"`
	IsPublished types.Bool     `tfsdk:"is_published"`
	Versions    []versionModel `tfsdk:"versions"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_versions"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `List versions of a Seqera Platform pipeline.

Wraps ` + "`GET /pipelines/{pipelineId}/versions`" + `. Use to discover ` + "`version_id`" + `s
for ` + "`seqera_pipeline_version`" + `, or to fan a launchpad over multiple revisions.`,
		Attributes: map[string]schema.Attribute{
			"pipeline_id": schema.Int64Attribute{
				Required:    true,
				Description: `Pipeline numeric identifier.`,
			},
			"workspace_id": schema.Int64Attribute{
				Required:    true,
				Description: `Workspace numeric identifier.`,
			},
			"is_published": schema.BoolAttribute{
				Optional:    true,
				Description: `Filter by publish status. Set to true for published versions only, false for drafts only. Omit to return both. Forwarded to the platform as the ?isPublished query parameter.`,
			},
			"versions": schema.ListNestedAttribute{
				Computed:    true,
				Description: `Pipeline versions, in the order returned by the API.`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                 schema.StringAttribute{Computed: true, Description: "Pipeline version string identifier."},
						"name":               schema.StringAttribute{Computed: true, Description: "Pipeline version name."},
						"is_default":         schema.BoolAttribute{Computed: true, Description: "Whether this version is the pipeline default."},
						"hash":               schema.StringAttribute{Computed: true, Description: "Server-side hash of the version's launch configuration."},
						"date_created":       schema.StringAttribute{Computed: true, Description: "RFC3339 timestamp when the version was created."},
						"last_updated":       schema.StringAttribute{Computed: true, Description: "RFC3339 timestamp when the version was last updated."},
						"creator_user_id":    schema.Int64Attribute{Computed: true, Description: "Numeric ID of the user who created the version."},
						"creator_user_name":  schema.StringAttribute{Computed: true, Description: "Username of the version creator."},
						"creator_first_name": schema.StringAttribute{Computed: true, Description: "First name of the version creator."},
						"creator_last_name":  schema.StringAttribute{Computed: true, Description: "Last name of the version creator."},
						"creator_avatar_url": schema.StringAttribute{Computed: true, Description: "Avatar URL of the version creator."},
					},
				},
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

	workspaceID := data.WorkspaceID.ValueInt64()
	listReq := operations.ListPipelineVersionsRequest{
		PipelineID:  data.PipelineID.ValueInt64(),
		WorkspaceID: &workspaceID,
	}
	if !data.IsPublished.IsNull() && !data.IsPublished.IsUnknown() {
		v := data.IsPublished.ValueBool()
		listReq.IsPublished = &v
	}
	res, err := d.client.PipelineVersions.ListPipelineVersions(ctx, listReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list pipeline versions", err.Error())
		return
	}
	if res.StatusCode != http.StatusOK || res.ListPipelineVersionsResponse == nil {
		resp.Diagnostics.AddError("Unexpected API response", common.DebugResponse(res.RawResponse))
		return
	}

	data.Versions = make([]versionModel, 0, len(res.ListPipelineVersionsResponse.Versions))
	for _, p := range res.ListPipelineVersionsResponse.Versions {
		if p.Version == nil {
			continue
		}
		v := p.Version
		vm := versionModel{
			ID:               types.StringPointerValue(v.ID),
			Name:             types.StringPointerValue(v.Name),
			IsDefault:        types.BoolPointerValue(v.IsDefault),
			Hash:             types.StringPointerValue(v.Hash),
			DateCreated:      rfc3339(v.DateCreated),
			LastUpdated:      rfc3339(v.LastUpdated),
			CreatorUserID:    types.Int64PointerValue(v.CreatorUserID),
			CreatorUserName:  types.StringPointerValue(v.CreatorUserName),
			CreatorFirstName: types.StringPointerValue(v.CreatorFirstName),
			CreatorLastName:  types.StringPointerValue(v.CreatorLastName),
			CreatorAvatarURL: types.StringPointerValue(v.CreatorAvatarURL),
		}
		data.Versions = append(data.Versions, vm)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func rfc3339(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}
