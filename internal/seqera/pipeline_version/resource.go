// Package pipeline_version provides the seqera_pipeline_version resource —
// a Terraform handle on an existing pipeline version that owns its `name`
// and `is_default` flag.
//
// The Seqera API has no endpoint to create or delete an individual version:
// versions are created server-side when a versionable field on the parent
// pipeline changes, and the audit trail is immutable. The only mutation
// primitive is PUT /pipelines/{pipelineId}/versions/{versionId}/manage,
// which sets `name` and `isDefault`. This resource is a Terraform binding
// for exactly that primitive — see the schema MarkdownDescription for
// user-facing semantics.
package pipeline_version

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
	"github.com/seqeralabs/terraform-provider-seqera/internal/seqera/common"
)

var _ resource.Resource = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *sdk.Seqera
}

type ResourceModel struct {
	PipelineID  types.Int64  `tfsdk:"pipeline_id"`
	WorkspaceID types.Int64  `tfsdk:"workspace_id"`
	VersionID   types.String `tfsdk:"version_id"`
	Name        types.String `tfsdk:"name"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Hash        types.String `tfsdk:"hash"`
	DateCreated types.String `tfsdk:"date_created"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_version"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Own the ` + "`name`" + ` and ` + "`is_default`" + ` flag of an existing Seqera Platform pipeline version.

Wraps ` + "`PUT /pipelines/{pipelineId}/versions/{versionId}/manage`" + `. Use with the
` + "`seqera_pipeline_versions`" + ` data source to discover ` + "`version_id`" + `:

` + "```hcl" + `
data "seqera_pipeline_versions" "hello" {
  pipeline_id  = seqera_pipeline.hello.pipeline_id
  workspace_id = var.workspace_id
}

resource "seqera_pipeline_version" "hello_v1" {
  pipeline_id  = seqera_pipeline.hello.pipeline_id
  workspace_id = var.workspace_id
  version_id   = [for v in data.seqera_pipeline_versions.hello.versions : v.id if v.name == "hello-1"][0]
  name         = "release-2024-Q4"
  is_default   = true
}
` + "```" + `

The Seqera API has no endpoint for creating or deleting individual versions
— they are server-created during pipeline create or when a versionable
field on ` + "`seqera_pipeline`" + ` changes. Use this resource to *manage* an
existing version (rename, promote, demote), not to create one.

` + "`terraform destroy`" + ` is a no-op: the API cannot unset ` + "`isDefault`" + ` on the
current default (it would leave zero defaults) and has no
DELETE-by-version-id. Destroying releases Terraform's ownership; the
version remains in the platform.`,
		Attributes: map[string]schema.Attribute{
			"pipeline_id": schema.Int64Attribute{
				Required:    true,
				Description: `Pipeline numeric identifier.`,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"workspace_id": schema.Int64Attribute{
				Required:    true,
				Description: `Workspace numeric identifier.`,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"version_id": schema.StringAttribute{
				Required:    true,
				Description: `Pipeline version string identifier owned by this resource.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `Display name for this pipeline version. Updated in place; renames do not create a new version.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: `Whether this version is the pipeline default. Setting true promotes this version and demotes any previous default. Cannot be set to false on the current default version without first promoting another (the platform refuses to leave a pipeline with zero defaults).`,
			},
			"hash": schema.StringAttribute{
				Computed:    true,
				Description: `Server-side hash of the version's launch configuration.`,
			},
			"date_created": schema.StringAttribute{
				Computed:    true,
				Description: `RFC3339 timestamp when the version was created.`,
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: `RFC3339 timestamp when the version was last updated.`,
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.Seqera)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Seqera, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

// manage calls PUT /manage. Always re-assert isDefault: the platform reads
// an omitted/false isDefault as "demote", which 409s on the current default.
func (r *Resource) manage(ctx context.Context, data ResourceModel) error {
	isDefault := data.IsDefault.ValueBool()
	body := shared.PipelineVersionManageRequest{IsDefault: &isDefault}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		body.Name = &name
	}
	workspaceID := data.WorkspaceID.ValueInt64()
	res, err := r.client.PipelineVersions.ManagePipelineVersion(ctx, operations.ManagePipelineVersionRequest{
		PipelineID:                   data.PipelineID.ValueInt64(),
		VersionID:                    data.VersionID.ValueString(),
		WorkspaceID:                  &workspaceID,
		PipelineVersionManageRequest: body,
	})
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d:\n%s", res.StatusCode, common.DebugResponse(res.RawResponse))
	}
	return nil
}

// refresh lists the parent pipeline's versions and copies the matching
// row's fields into data. Returns false if the version is gone.
// /manage returns 204 with no body, so a list-and-find is the only way to
// observe hash/timestamps after a write — there is no GET-by-version-id.
func (r *Resource) refresh(ctx context.Context, data *ResourceModel) (bool, error) {
	pipelineID := data.PipelineID.ValueInt64()
	workspaceID := data.WorkspaceID.ValueInt64()
	versionID := data.VersionID.ValueString()
	res, err := r.client.PipelineVersions.ListPipelineVersions(ctx, operations.ListPipelineVersionsRequest{
		PipelineID:  pipelineID,
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode != http.StatusOK || res.ListPipelineVersionsResponse == nil {
		return false, fmt.Errorf("unexpected status %d listing versions:\n%s", res.StatusCode, common.DebugResponse(res.RawResponse))
	}
	for _, p := range res.ListPipelineVersionsResponse.Versions {
		v := p.Version
		if v == nil || v.ID == nil || *v.ID != versionID {
			continue
		}
		data.Name = types.StringPointerValue(v.Name)
		data.IsDefault = types.BoolPointerValue(v.IsDefault)
		data.Hash = types.StringPointerValue(v.Hash)
		data.DateCreated = rfc3339(v.DateCreated)
		data.LastUpdated = rfc3339(v.LastUpdated)
		return true, nil
	}
	return false, nil
}

// applyAndRefresh manages the version then reads back observable fields.
func (r *Resource) applyAndRefresh(ctx context.Context, data *ResourceModel) (found bool, err error) {
	if err := r.manage(ctx, *data); err != nil {
		return false, fmt.Errorf("failed to manage pipeline version: %w", err)
	}
	return r.refresh(ctx, data)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := r.applyAndRefresh(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pipeline version assignment", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Pipeline version not found after manage",
			fmt.Sprintf("version_id %q is not present in pipeline %d's version list", data.VersionID.ValueString(), data.PipelineID.ValueInt64()),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := r.refresh(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list pipeline versions", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip the network call when name + is_default already match state.
	if plan.Name.Equal(state.Name) && plan.IsDefault.Equal(state.IsDefault) {
		if _, err := r.refresh(ctx, &plan); err != nil {
			resp.Diagnostics.AddError("Failed to read pipeline version", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if _, err := r.applyAndRefresh(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to update pipeline version assignment", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op. The API cannot unset isDefault on the current default
// and has no DELETE-by-version-id; destroying releases Terraform's
// ownership and leaves the platform state untouched.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func rfc3339(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}
