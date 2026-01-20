// Package organization_member provides the seqera_organization_member resource.
package organization_member

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
	"github.com/seqeralabs/terraform-provider-seqera/internal/seqera/common"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *sdk.Seqera
}

type ResourceModel struct {
	OrgID     types.Int64  `tfsdk:"org_id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	MemberID  types.Int64  `tfsdk:"member_id"`
	UserID    types.Int64  `tfsdk:"user_id"`
	UserName  types.String `tfsdk:"user_name"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Avatar    types.String `tfsdk:"avatar"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manage organization members in Seqera Platform.

Organization members are users who have access to an organization. Each member
has a role that determines their permissions within the organization.

Available roles:
- owner: Full control over the organization
- member: Standard member access (default)
- collaborator: Limited collaboration access

Import format: org_id/email (e.g., "12345/user@example.com")
`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: `Organization numeric identifier.`,
			},
			"email": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: `Email address of the user to add as an organization member.`,
			},
			"role": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("member"),
				MarkdownDescription: `Role of the member. Valid values: owner, member, collaborator. Defaults to "member".`,
				Validators: []validator.String{
					stringvalidator.OneOf("owner", "member", "collaborator"),
				},
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
			"avatar": schema.StringAttribute{
				Computed:    true,
				Description: `Avatar URL of the member.`,
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

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the member
	email := data.Email.ValueString()
	createRes, err := r.client.Orgs.CreateOrganizationMember(ctx, operations.CreateOrganizationMemberRequest{
		OrgID:            data.OrgID.ValueInt64(),
		AddMemberRequest: shared.AddMemberRequest{User: &email},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create organization member", err.Error())
		return
	}
	if createRes.StatusCode == 409 {
		resp.Diagnostics.AddError("Resource Already Exists", "The user is already a member of this organization.")
		return
	}
	if createRes.StatusCode != 200 || createRes.AddMemberResponse == nil || createRes.AddMemberResponse.Member == nil {
		resp.Diagnostics.AddError("Unexpected API response", common.DebugResponse(createRes.RawResponse))
		return
	}

	// Store member details
	member := createRes.AddMemberResponse.Member
	r.refreshFromMember(&data, member)

	// Update role if not default
	desiredRole := data.Role.ValueString()
	if desiredRole != "" && desiredRole != "member" {
		role := shared.OrgRole(desiredRole)
		updateRes, err := r.client.Orgs.UpdateOrganizationMemberRole(ctx, operations.UpdateOrganizationMemberRoleRequest{
			OrgID:                   data.OrgID.ValueInt64(),
			MemberID:                data.MemberID.ValueInt64(),
			UpdateMemberRoleRequest: shared.UpdateMemberRoleRequest{Role: &role},
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to update member role", err.Error())
			return
		}
		if updateRes.StatusCode != 204 {
			resp.Diagnostics.AddError("Failed to update member role", common.DebugResponse(updateRes.RawResponse))
			return
		}
		data.Role = types.StringValue(desiredRole)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.findMember(ctx, data.OrgID.ValueInt64(), data.Email.ValueString(), data.MemberID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read organization member", err.Error())
		return
	}
	if member == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.refreshFromMember(&data, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ResourceModel
	var state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve member_id from state
	data.MemberID = state.MemberID

	// Update role
	role := shared.OrgRole(data.Role.ValueString())
	updateRes, err := r.client.Orgs.UpdateOrganizationMemberRole(ctx, operations.UpdateOrganizationMemberRoleRequest{
		OrgID:                   data.OrgID.ValueInt64(),
		MemberID:                data.MemberID.ValueInt64(),
		UpdateMemberRoleRequest: shared.UpdateMemberRoleRequest{Role: &role},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update member role", err.Error())
		return
	}
	if updateRes.StatusCode != 204 {
		resp.Diagnostics.AddError("Failed to update member role", common.DebugResponse(updateRes.RawResponse))
		return
	}

	// Refresh from API
	member, err := r.findMember(ctx, data.OrgID.ValueInt64(), data.Email.ValueString(), data.MemberID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to refresh member", err.Error())
		return
	}
	if member != nil {
		r.refreshFromMember(&data, member)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteRes, err := r.client.Orgs.DeleteOrganizationMember(ctx, operations.DeleteOrganizationMemberRequest{
		OrgID:    data.OrgID.ValueInt64(),
		MemberID: data.MemberID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete organization member", err.Error())
		return
	}
	if deleteRes.StatusCode != 204 && deleteRes.StatusCode != 404 {
		resp.Diagnostics.AddError("Failed to delete organization member", common.DebugResponse(deleteRes.RawResponse))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: org_id/email
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: org_id/email, got: %s", req.ID),
		)
		return
	}

	orgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid org_id",
			fmt.Sprintf("org_id must be a number, got: %s", parts[0]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), parts[1])...)
}

// findMember searches for a member by email or member_id.
// Note: The API does not support pagination. Large organizations may not return all members.
func (r *Resource) findMember(ctx context.Context, orgID int64, email string, memberID int64) (*shared.MemberDbDto, error) {
	listRes, err := r.client.Orgs.ListOrganizationMembers(ctx, operations.ListOrganizationMembersRequest{
		OrgID:  orgID,
		Search: &email,
	})
	if err != nil {
		return nil, err
	}
	if listRes.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d listing organization members", listRes.StatusCode)
	}
	if listRes.ListMembersResponse == nil {
		return nil, fmt.Errorf("empty response listing organization members")
	}

	for i := range listRes.ListMembersResponse.Members {
		m := &listRes.ListMembersResponse.Members[i]
		if (m.Email != nil && *m.Email == email) || (m.MemberID != nil && *m.MemberID == memberID) {
			return m, nil
		}
	}
	return nil, nil
}

// refreshFromMember updates the ResourceModel from API response.
func (r *Resource) refreshFromMember(data *ResourceModel, member *shared.MemberDbDto) {
	data.MemberID = types.Int64PointerValue(member.MemberID)
	data.UserID = types.Int64PointerValue(member.UserID)
	data.UserName = types.StringPointerValue(member.UserName)
	data.Email = types.StringPointerValue(member.Email)
	data.FirstName = types.StringPointerValue(member.FirstName)
	data.LastName = types.StringPointerValue(member.LastName)
	data.Avatar = types.StringPointerValue(member.Avatar)
	if member.Role != nil {
		data.Role = types.StringValue(string(*member.Role))
	}
}
