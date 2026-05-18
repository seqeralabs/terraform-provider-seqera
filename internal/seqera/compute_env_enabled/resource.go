// Package compute_env_enabled provides the seqera_compute_env_enabled
// resource. It lives outside the generated typed CE resources because the
// platform exposes enable / disable as two POST-without-body endpoints
// (POST /compute-envs/{id}/enable, POST /compute-envs/{id}/disable), which
// don't fit Speakeasy's CRUD generation. See the schema MarkdownDescription
// for user-facing semantics.
package compute_env_enabled

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
	"github.com/seqeralabs/terraform-provider-seqera/internal/seqera/common"
)

const disabledStatus = "DISABLED"

var _ resource.Resource = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *sdk.Seqera
}

type ResourceModel struct {
	ComputeEnvID types.String `tfsdk:"compute_env_id"`
	WorkspaceID  types.Int64  `tfsdk:"workspace_id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Status       types.String `tfsdk:"status"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_env_enabled"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Own the enable/disable state of an existing Seqera compute environment.

Wraps the ` + "`POST /compute-envs/{id}/enable`" + ` and ` + "`POST /compute-envs/{id}/disable`" + ` endpoints, available on Seqera Enterprise v26.1+ and Seqera Platform Cloud.

A disabled compute environment cannot launch new workflows but remains visible and intact; re-enable to resume launching. The provider treats enable / disable as a separate assignment-style resource because the API endpoints don't fit Speakeasy's generated CRUD shape (no PATCH on the parent CE, no body on the enable / disable POSTs).

` + "```hcl" + `
resource "seqera_compute_env_enabled" "freeze_prod" {
  compute_env_id = seqera_aws_batch_ce.prod.compute_env.id
  workspace_id   = seqera_aws_batch_ce.prod.workspace_id
  enabled        = false
}
` + "```" + `

` + "`terraform destroy`" + ` is a no-op: the CE retains whatever enable / disable state was last applied. Removing the resource releases Terraform's ownership of the state without flipping it.`,
		Attributes: map[string]schema.Attribute{
			"compute_env_id": schema.StringAttribute{
				Required:    true,
				Description: `Compute environment string identifier.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: `Desired enable state. true → POST /enable; false → POST /disable.`,
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: `Current CE status as reported by the platform (e.g. AVAILABLE, DISABLED, ERRORED). Surfaced for observability.`,
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

func (r *Resource) apply(ctx context.Context, data ResourceModel) error {
	ceID := data.ComputeEnvID.ValueString()
	workspaceID := data.WorkspaceID.ValueInt64()
	enabled := data.Enabled.ValueBool()

	var statusCode int
	var raw *http.Response
	var err error
	if enabled {
		var res *operations.EnableComputeEnvResponse
		res, err = r.client.ComputeEnvs.EnableComputeEnv(ctx, operations.EnableComputeEnvRequest{
			ComputeEnvID: ceID,
			WorkspaceID:  &workspaceID,
		})
		if res != nil {
			statusCode, raw = res.StatusCode, res.RawResponse
		}
	} else {
		var res *operations.DisableComputeEnvResponse
		res, err = r.client.ComputeEnvs.DisableComputeEnv(ctx, operations.DisableComputeEnvRequest{
			ComputeEnvID: ceID,
			WorkspaceID:  &workspaceID,
		})
		if res != nil {
			statusCode, raw = res.StatusCode, res.RawResponse
		}
	}
	if err != nil {
		return err
	}
	verb := "disabling"
	if enabled {
		verb = "enabling"
	}
	if statusCode != http.StatusNoContent {
		return common.UnexpectedStatusErr(verb+" compute env", raw)
	}
	return nil
}

// refresh reads the CE and populates `enabled` + `status` from the server.
// Returns false if the CE is missing — the assignment is then removed from state.
func (r *Resource) refresh(ctx context.Context, data *ResourceModel) (bool, error) {
	ceID := data.ComputeEnvID.ValueString()
	workspaceID := data.WorkspaceID.ValueInt64()
	res, err := r.client.ComputeEnvs.DescribeComputeEnv(ctx, operations.DescribeComputeEnvRequest{
		ComputeEnvID: ceID,
		WorkspaceID:  workspaceID,
	})
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		return false, nil
	}
	if res.StatusCode != http.StatusOK || res.DescribeComputeEnvResponse == nil || res.DescribeComputeEnvResponse.ComputeEnv == nil {
		return false, common.UnexpectedStatusErr("describing compute env", res.RawResponse)
	}
	ce := res.DescribeComputeEnvResponse.ComputeEnv
	if ce.Status == nil {
		// `Status` is documented as always-present on a live CE; if the
		// server returns null treat it as a missing CE rather than guessing.
		return false, nil
	}
	data.Status = types.StringValue(string(*ce.Status))
	data.Enabled = types.BoolValue(string(*ce.Status) != disabledStatus)
	return true, nil
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to set compute env enabled state", err.Error())
		return
	}
	if _, err := r.refresh(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to read compute env after apply", err.Error())
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
		resp.Diagnostics.AddError("Failed to describe compute env", err.Error())
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
	if plan.Enabled.Equal(state.Enabled) {
		if _, err := r.refresh(ctx, &plan); err != nil {
			resp.Diagnostics.AddError("Failed to describe compute env", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	if err := r.apply(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Failed to set compute env enabled state", err.Error())
		return
	}
	if _, err := r.refresh(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to read compute env after apply", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op. Destroying the resource releases Terraform's ownership
// of the enable / disable state; the CE keeps whatever state was last applied.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
