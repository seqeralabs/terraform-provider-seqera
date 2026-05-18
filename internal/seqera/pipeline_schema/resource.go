// Package pipeline_schema provides the seqera_pipeline_schema resource.
//
// The /pipeline-schemas endpoint only supports POST — there is no read,
// update, or delete by schema id. This resource wraps that single endpoint
// and exposes the server-assigned `id` so users can reference it from
// `seqera_pipeline.launch.pipeline_schema_id`.
//
// Because the API is write-only:
//   - schema_content changes force resource replacement (POST returns a new
//     id every call; there is no in-place update).
//   - Read is trusted from state — there is no GET-by-schema-id to reconcile
//     against. The row is immutable server-side once created, so drift
//     between state and server is not expected.
//   - Delete is a no-op; the schema row is left orphaned in the backend.
package pipeline_schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	WorkspaceID   types.Int64  `tfsdk:"workspace_id"`
	SchemaContent types.String `tfsdk:"schema_content"`
	ID            types.Int64  `tfsdk:"id"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_schema"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Create a Nextflow pipeline schema in a Seqera workspace and expose its server-assigned id.

Reference the returned id from ` + "`seqera_pipeline.launch.pipeline_schema_id`" + ` to bind the schema to a pipeline:

` + "```hcl" + `
resource "seqera_pipeline_schema" "rnaseq" {
  workspace_id   = var.workspace_id
  schema_content = file("${path.module}/schema.json")
}

resource "seqera_pipeline" "rnaseq" {
  workspace_id = var.workspace_id
  name         = "rnaseq"
  launch = {
    pipeline           = "https://github.com/nf-core/rnaseq"
    compute_env_id     = seqera_compute_env.aws.compute_env.id
    pipeline_schema_id = seqera_pipeline_schema.rnaseq.id
  }

  depends_on = [seqera_pipeline_schema.rnaseq]
}
` + "```" + `

The Seqera API has no read, update, or delete path for a schema by id, so
this resource is effectively write-once: changes to schema_content force
replacement, delete is a no-op, and the previous schema row is left
orphaned server-side.
`,
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.Int64Attribute{
				Required:    true,
				Description: `Workspace numeric identifier the schema is created in.`,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"schema_content": schema.StringAttribute{
				Required:    true,
				Description: `Raw Nextflow pipeline schema JSON. Changes force resource replacement because the API has no update endpoint.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: `Server-assigned pipeline schema id. Reference this from seqera_pipeline.launch.pipeline_schema_id.`,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := common.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if client != nil {
		r.client = client
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := data.WorkspaceID.ValueInt64()
	content := data.SchemaContent.ValueString()

	res, err := r.client.PipelineSchemas.CreatePipelineSchema(ctx, operations.CreatePipelineSchemaRequest{
		WorkspaceID: &workspaceID,
		CreatePipelineSchemaRequest: shared.CreatePipelineSchemaRequest{
			Content: &content,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pipeline schema", err.Error())
		return
	}
	if res.StatusCode != 200 ||
		res.CreatePipelineSchemaResponse == nil ||
		res.CreatePipelineSchemaResponse.PipelineSchema == nil ||
		res.CreatePipelineSchemaResponse.PipelineSchema.ID == nil {
		common.AddUnexpectedStatus(&resp.Diagnostics, "creating pipeline schema", res.RawResponse)
		return
	}

	data.ID = types.Int64PointerValue(res.CreatePipelineSchemaResponse.PipelineSchema.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is a no-op: the API has no GET-by-schema-id endpoint, so we trust
// state as the source of truth. Schema rows are immutable server-side, so
// drift between state and server is not expected.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is unreachable: every user-facing field has RequiresReplace, so
// terraform will destroy+create rather than call Update. Implemented to
// satisfy the resource.Resource interface.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete is a no-op: the API has no DELETE-by-schema-id endpoint. The row
// is left orphaned in the backend.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
