// Package team_member provides the seqera_team_member resource.
package team_member

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	TeamID    types.Int64  `tfsdk:"team_id"`
	MemberID  types.Int64  `tfsdk:"member_id"`
	Email     types.String `tfsdk:"email"`
	UserID    types.Int64  `tfsdk:"user_id"`
	UserName  types.String `tfsdk:"user_name"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Avatar    types.String `tfsdk:"avatar"`
	Role      types.String `tfsdk:"role"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_member"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manage team members in Seqera Platform.

Team members are organization members who have been added to a specific team.
Teams can be used to organize users and grant workspace access collectively.

Note: The user must already be a member of the organization before they can
be added to a team.

Import format: org_id/team_id/email (e.g., "12345/67890/user@example.com")
`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: `Organization numeric identifier.`,
			},
			"team_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: `Team numeric identifier.`,
			},
			"member_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
					int64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: `Organization member ID to add to the team. Specify either member_id or email but not both.`,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("email"),
					}...),
					int64validator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("email"),
					}...),
				},
			},
			"email": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: `Email address or username of the user to add to the team. Specify either member_id or email but not both.`,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("member_id"),
					}...),
					stringvalidator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("member_id"),
					}...),
				},
			},
			"user_id": schema.Int64Attribute{
				Computed:    true,
				Description: `User numeric identifier.`,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"user_name": schema.StringAttribute{
				Computed:    true,
				Description: `Username of the member.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"first_name": schema.StringAttribute{
				Computed:    true,
				Description: `First name of the member.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_name": schema.StringAttribute{
				Computed:    true,
				Description: `Last name of the member.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"avatar": schema.StringAttribute{
				Computed:    true,
				Description: `Avatar URL of the member.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Computed:    true,
				Description: `Organization role of the member.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	// Build request - either member_id or email must be specified
	// The API only supports userNameOrEmail, so if member_id is provided,
	// we need to look up the email first
	var email string
	if !data.Email.IsNull() && !data.Email.IsUnknown() {
		email = data.Email.ValueString()
	} else if !data.MemberID.IsNull() && !data.MemberID.IsUnknown() {
		// Look up the member to get their email
		member, err := r.findOrgMember(ctx, data.OrgID.ValueInt64(), data.MemberID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Failed to find organization member", err.Error())
			return
		}
		if member == nil {
			resp.Diagnostics.AddError("Member Not Found", fmt.Sprintf("No organization member found with ID %d", data.MemberID.ValueInt64()))
			return
		}
		if member.Email != nil {
			email = *member.Email
		} else {
			resp.Diagnostics.AddError("Member Email Not Found", "The organization member does not have an email address")
			return
		}
	} else {
		resp.Diagnostics.AddError("Invalid Configuration", "Either member_id or email must be specified.")
		return
	}

	createRes, err := r.client.Teams.CreateOrganizationTeamMember(ctx, operations.CreateOrganizationTeamMemberRequest{
		OrgID:                   data.OrgID.ValueInt64(),
		TeamID:                  data.TeamID.ValueInt64(),
		CreateTeamMemberRequest: shared.CreateTeamMemberRequest{UserNameOrEmail: &email},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create team member", err.Error())
		return
	}
	if createRes.StatusCode == 409 {
		resp.Diagnostics.AddError("Resource Already Exists", "The user is already a member of this team.")
		return
	}
	if createRes.StatusCode != 200 || createRes.AddTeamMemberResponse == nil || createRes.AddTeamMemberResponse.Member == nil {
		resp.Diagnostics.AddError("Unexpected API response", common.DebugResponse(createRes.RawResponse))
		return
	}

	r.refreshFromMember(&data, createRes.AddTeamMemberResponse.Member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use empty email since we have member_id from state - triggers ID-based lookup optimization
	member, err := r.findMember(ctx, data.OrgID.ValueInt64(), data.TeamID.ValueInt64(), "", data.MemberID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read team member", err.Error())
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
	// All configurable attributes require replacement, so this should never be called
	resp.Diagnostics.AddError("Update Not Supported", "Team member resources cannot be updated in place.")
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteRes, err := r.client.Teams.DeleteOrganizationTeamMember(ctx, operations.DeleteOrganizationTeamMemberRequest{
		OrgID:    data.OrgID.ValueInt64(),
		TeamID:   data.TeamID.ValueInt64(),
		MemberID: data.MemberID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete team member", err.Error())
		return
	}
	if deleteRes.StatusCode != 204 && deleteRes.StatusCode != 404 {
		resp.Diagnostics.AddError("Failed to delete team member", common.DebugResponse(deleteRes.RawResponse))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: org_id/team_id/email
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: org_id/team_id/email, got: %s", req.ID),
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

	teamID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid team_id",
			fmt.Sprintf("team_id must be a number, got: %s", parts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), teamID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), parts[2])...)
}

// findOrgMember looks up an organization member by member_id.
// Note: The API does not support pagination. Large organizations may not return all members.
func (r *Resource) findOrgMember(ctx context.Context, orgID, memberID int64) (*shared.MemberDbDto, error) {
	listRes, err := r.client.Orgs.ListOrganizationMembers(ctx, operations.ListOrganizationMembersRequest{
		OrgID: orgID,
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
		if m.MemberID != nil && *m.MemberID == memberID {
			return m, nil
		}
	}
	return nil, nil
}

// findMember searches for a team member by email or member_id.
// When member_id is provided (non-zero), it searches without email filter for better performance.
// Note: The API does not support pagination. Large teams may not return all members.
func (r *Resource) findMember(ctx context.Context, orgID, teamID int64, email string, memberID int64) (*shared.MemberDbDto, error) {
	// If we have a member_id, don't use email search - just get all members and filter by ID
	// This avoids email lookup latency and is more efficient
	var searchParam *string
	if memberID > 0 {
		// Use empty search when we have member_id - we'll filter by ID in the loop
		emptySearch := ""
		searchParam = &emptySearch
	} else {
		// Only use email search if we don't have member_id (e.g., during import)
		searchParam = &email
	}

	listRes, err := r.client.Teams.ListOrganizationTeamMembers(ctx, operations.ListOrganizationTeamMembersRequest{
		OrgID:  orgID,
		TeamID: teamID,
		Search: searchParam,
	})
	if err != nil {
		return nil, err
	}
	if listRes.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d listing team members", listRes.StatusCode)
	}
	if listRes.ListMembersResponse == nil {
		return nil, fmt.Errorf("empty response listing team members")
	}

	for i := range listRes.ListMembersResponse.Members {
		m := &listRes.ListMembersResponse.Members[i]
		// Prefer matching by member_id if available, fallback to email
		if memberID > 0 && m.MemberID != nil && *m.MemberID == memberID {
			return m, nil
		}
		if memberID == 0 && m.Email != nil && *m.Email == email {
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
