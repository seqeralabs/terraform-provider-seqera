// Package workspace_participant provides the seqera_workspace_participant resource.
package workspace_participant

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
	OrgID         types.Int64  `tfsdk:"org_id"`
	WorkspaceID   types.Int64  `tfsdk:"workspace_id"`
	MemberID      types.Int64  `tfsdk:"member_id"`
	TeamID        types.Int64  `tfsdk:"team_id"`
	Email         types.String `tfsdk:"email"`
	Role          types.String `tfsdk:"role"`
	ParticipantID types.Int64  `tfsdk:"participant_id"`
	UserName      types.String `tfsdk:"user_name"`
	FirstName     types.String `tfsdk:"first_name"`
	LastName      types.String `tfsdk:"last_name"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_participant"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manage workspace participants in Seqera Platform.

Workspace participants control access to workspace resources. Each participant
can be an individual organization member (via member_id or email) or an entire
team (via team_id) with a specific role that determines their permissions within
the workspace.

Note: When using email, the lookup to member_id happens once during resource
creation and the participant_id is stored in state for subsequent operations.

Available roles:
- owner: Full control over the workspace
- admin: Administrative access, can manage participants
- maintain: Can modify pipelines and compute environments
- launch: Can launch pipelines
- view: Read-only access (default)

Import formats:
- org_id/workspace_id/email (e.g., "12345/67890/user@example.com")
- org_id/workspace_id/team:team_id (e.g., "12345/67890/team:7405043533023")
- org_id/workspace_id/member:member_id (e.g., "12345/67890/member:98765")
`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: `Organization numeric identifier.`,
			},
			"workspace_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: `Workspace numeric identifier.`,
			},
			"member_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: `Organization member ID to add as a workspace participant. Specify either member_id, team_id, or email but not multiple.`,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("team_id"),
						path.MatchRoot("email"),
					}...),
					int64validator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("team_id"),
						path.MatchRoot("email"),
					}...),
				},
			},
			"team_id": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: `Team ID to add as workspace participants. All team members will be granted access. Specify either member_id, team_id, or email but not multiple.`,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.Expressions{
						path.MatchRoot("member_id"),
						path.MatchRoot("email"),
					}...),
					int64validator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("member_id"),
						path.MatchRoot("email"),
					}...),
				},
			},
			"email": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: `Email address of the user to add as a workspace participant. Specify either member_id, team_id, or email but not multiple. Email lookup happens once during creation.`,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("member_id"),
						path.MatchRoot("team_id"),
					}...),
					stringvalidator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("member_id"),
						path.MatchRoot("team_id"),
					}...),
				},
			},
			"role": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("view"),
				MarkdownDescription: `Role of the participant. Valid values: owner, admin, maintain, launch, view. Defaults to "view".`,
				Validators: []validator.String{
					stringvalidator.OneOf("owner", "admin", "maintain", "launch", "view"),
				},
			},
			"participant_id": schema.Int64Attribute{
				Computed:    true,
				Description: `Participant numeric identifier.`,
			},
			"user_name": schema.StringAttribute{
				Computed:    true,
				Description: `Username of the participant.`,
			},
			"first_name": schema.StringAttribute{
				Computed:    true,
				Description: `First name of the participant.`,
			},
			"last_name": schema.StringAttribute{
				Computed:    true,
				Description: `Last name of the participant.`,
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

	// Build request - either member_id, team_id, or email must be specified
	addReq := shared.AddParticipantRequest{}
	if !data.MemberID.IsNull() && !data.MemberID.IsUnknown() {
		memberID := data.MemberID.ValueInt64()
		addReq.MemberID = &memberID
	} else if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
		teamID := data.TeamID.ValueInt64()
		addReq.TeamID = &teamID
	} else if !data.Email.IsNull() && !data.Email.IsUnknown() {
		email := data.Email.ValueString()
		addReq.UserNameOrEmail = &email
	} else {
		resp.Diagnostics.AddError("Invalid Configuration", "Either member_id, team_id, or email must be specified.")
		return
	}

	createRes, err := r.client.Workspaces.CreateWorkspaceParticipant(ctx, operations.CreateWorkspaceParticipantRequest{
		OrgID:                 data.OrgID.ValueInt64(),
		WorkspaceID:           data.WorkspaceID.ValueInt64(),
		AddParticipantRequest: addReq,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create workspace participant", err.Error())
		return
	}
	if createRes.StatusCode == 409 {
		// Build helpful import message based on what was provided
		var importMsg string
		if !data.Email.IsNull() && !data.Email.IsUnknown() {
			email := data.Email.ValueString()
			importMsg = fmt.Sprintf(
				"This participant already exists in the workspace. Import it into Terraform state:\n\n"+
					"  terraform import seqera_workspace_participant.<name> '%d/%d/%s'\n\n"+
					"Or remove it from the workspace before managing it with Terraform.",
				data.OrgID.ValueInt64(),
				data.WorkspaceID.ValueInt64(),
				email,
			)
		} else if !data.TeamID.IsNull() && !data.TeamID.IsUnknown() {
			teamID := data.TeamID.ValueInt64()
			importMsg = fmt.Sprintf(
				"This team is already a participant in the workspace. Import it into Terraform state:\n\n"+
					"  terraform import seqera_workspace_participant.<name> '%d/%d/team:%d'\n\n"+
					"Or remove the team from the workspace before managing it with Terraform.",
				data.OrgID.ValueInt64(),
				data.WorkspaceID.ValueInt64(),
				teamID,
			)
		} else {
			// member_id case
			memberID := data.MemberID.ValueInt64()
			importMsg = fmt.Sprintf(
				"This participant already exists in the workspace. Import it into Terraform state:\n\n"+
					"  terraform import seqera_workspace_participant.<name> '%d/%d/member:%d'\n\n"+
					"Or remove it from the workspace before managing it with Terraform.",
				data.OrgID.ValueInt64(),
				data.WorkspaceID.ValueInt64(),
				memberID,
			)
		}

		resp.Diagnostics.AddError("Resource Already Exists", importMsg)
		return
	}
	if createRes.StatusCode != 200 || createRes.AddParticipantResponse == nil || createRes.AddParticipantResponse.Participant == nil {
		resp.Diagnostics.AddError("Unexpected API response", common.DebugResponse(createRes.RawResponse))
		return
	}

	participant := createRes.AddParticipantResponse.Participant
	data.ParticipantID = types.Int64PointerValue(participant.ParticipantID)

	// For team participants, explicitly set user-specific fields to null
	// For user participants, populate from API response
	isTeamParticipant := !data.TeamID.IsNull() && !data.TeamID.IsUnknown()
	if isTeamParticipant {
		// Team participants don't have individual user data
		data.MemberID = types.Int64Null()
		data.Email = types.StringNull()
		data.UserName = types.StringNull()
		data.FirstName = types.StringNull()
		data.LastName = types.StringNull()
	} else {
		// User participants have individual user data
		data.MemberID = types.Int64PointerValue(participant.MemberID)
		data.Email = types.StringPointerValue(participant.Email)
	}

	// Update role if not default
	desiredRole := data.Role.ValueString()
	if desiredRole != "" && desiredRole != "view" {
		role := desiredRole
		updateRes, err := r.client.Workspaces.UpdateWorkspaceParticipantRole(ctx, operations.UpdateWorkspaceParticipantRoleRequest{
			OrgID:                        data.OrgID.ValueInt64(),
			WorkspaceID:                  data.WorkspaceID.ValueInt64(),
			ParticipantID:                data.ParticipantID.ValueInt64(),
			UpdateParticipantRoleRequest: shared.UpdateParticipantRoleRequest{Role: &role},
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to update participant role", err.Error())
			return
		}
		if updateRes.StatusCode != 204 {
			resp.Diagnostics.AddError("Failed to update participant role", common.DebugResponse(updateRes.RawResponse))
			return
		}
	}

	// Refresh from list to get all computed fields (skip for team participants)
	if !isTeamParticipant {
		emailSearch := data.Email.ValueString()
		p, err := r.findParticipant(ctx, data.OrgID.ValueInt64(), data.WorkspaceID.ValueInt64(), emailSearch, data.ParticipantID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Failed to refresh participant", err.Error())
			return
		}
		if p != nil {
			r.refreshFromParticipant(&data, p)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Search by participant_id (works for both user and team participants)
	emailSearch := ""
	if !data.Email.IsNull() && !data.Email.IsUnknown() {
		emailSearch = data.Email.ValueString()
	}

	participant, err := r.findParticipant(ctx, data.OrgID.ValueInt64(), data.WorkspaceID.ValueInt64(), emailSearch, data.ParticipantID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read workspace participant", err.Error())
		return
	}
	if participant == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Only refresh member details for user participants (not team participants)
	isTeamParticipant := !data.TeamID.IsNull() && !data.TeamID.IsUnknown()
	if isTeamParticipant {
		// Team participants don't have individual user data
		data.MemberID = types.Int64Null()
		data.Email = types.StringNull()
		data.UserName = types.StringNull()
		data.FirstName = types.StringNull()
		data.LastName = types.StringNull()
	} else {
		r.refreshFromParticipant(&data, participant)
	}

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

	// Preserve participant_id from state
	data.ParticipantID = state.ParticipantID

	// Update role
	role := data.Role.ValueString()
	updateRes, err := r.client.Workspaces.UpdateWorkspaceParticipantRole(ctx, operations.UpdateWorkspaceParticipantRoleRequest{
		OrgID:                        data.OrgID.ValueInt64(),
		WorkspaceID:                  data.WorkspaceID.ValueInt64(),
		ParticipantID:                data.ParticipantID.ValueInt64(),
		UpdateParticipantRoleRequest: shared.UpdateParticipantRoleRequest{Role: &role},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update participant role", err.Error())
		return
	}
	if updateRes.StatusCode != 204 {
		resp.Diagnostics.AddError("Failed to update participant role", common.DebugResponse(updateRes.RawResponse))
		return
	}

	// Refresh from API
	isTeamParticipant := !data.TeamID.IsNull() && !data.TeamID.IsUnknown()
	if isTeamParticipant {
		// Team participants don't have individual user data
		data.MemberID = types.Int64Null()
		data.Email = types.StringNull()
		data.UserName = types.StringNull()
		data.FirstName = types.StringNull()
		data.LastName = types.StringNull()
	} else {
		emailSearch := ""
		if !data.Email.IsNull() && !data.Email.IsUnknown() {
			emailSearch = data.Email.ValueString()
		}

		participant, err := r.findParticipant(ctx, data.OrgID.ValueInt64(), data.WorkspaceID.ValueInt64(), emailSearch, data.ParticipantID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Failed to refresh participant", err.Error())
			return
		}
		if participant != nil {
			r.refreshFromParticipant(&data, participant)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteRes, err := r.client.Workspaces.DeleteWorkspaceParticipant(ctx, operations.DeleteWorkspaceParticipantRequest{
		OrgID:         data.OrgID.ValueInt64(),
		WorkspaceID:   data.WorkspaceID.ValueInt64(),
		ParticipantID: data.ParticipantID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete workspace participant", err.Error())
		return
	}
	if deleteRes.StatusCode != 204 && deleteRes.StatusCode != 404 {
		resp.Diagnostics.AddError("Failed to delete workspace participant", common.DebugResponse(deleteRes.RawResponse))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import formats:
	// - org_id/workspace_id/email (e.g., "12345/67890/user@example.com")
	// - org_id/workspace_id/team:team_id (e.g., "12345/67890/team:7405043533023")
	// - org_id/workspace_id/member:member_id (e.g., "12345/67890/member:98765")
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: org_id/workspace_id/email or org_id/workspace_id/team:team_id or org_id/workspace_id/member:member_id, got: %s", req.ID),
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

	workspaceID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid workspace_id",
			fmt.Sprintf("workspace_id must be a number, got: %s", parts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)

	// Determine import type based on third part
	identifier := parts[2]

	if strings.HasPrefix(identifier, "team:") {
		// Import by team_id
		teamIDStr := strings.TrimPrefix(identifier, "team:")
		teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid team_id",
				fmt.Sprintf("team_id must be a number, got: %s", teamIDStr),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), teamID)...)
	} else if strings.HasPrefix(identifier, "member:") {
		// Import by member_id
		memberIDStr := strings.TrimPrefix(identifier, "member:")
		memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid member_id",
				fmt.Sprintf("member_id must be a number, got: %s", memberIDStr),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_id"), memberID)...)
	} else {
		// Import by email (default)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), identifier)...)
	}
}

// findParticipant searches for a participant by email or participant_id.
// Note: The API does not support pagination. Large workspaces may not return all participants.
func (r *Resource) findParticipant(ctx context.Context, orgID, workspaceID int64, email string, participantID int64) (*shared.ParticipantResponseDto, error) {
	listRes, err := r.client.Workspaces.ListWorkspaceParticipants(ctx, operations.ListWorkspaceParticipantsRequest{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Search:      &email,
	})
	if err != nil {
		return nil, err
	}
	if listRes.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d listing workspace participants", listRes.StatusCode)
	}
	if listRes.ListParticipantsResponse == nil {
		return nil, fmt.Errorf("empty response listing workspace participants")
	}

	for i := range listRes.ListParticipantsResponse.Participants {
		p := &listRes.ListParticipantsResponse.Participants[i]
		if (p.Email != nil && *p.Email == email) || (p.ParticipantID != nil && *p.ParticipantID == participantID) {
			return p, nil
		}
	}
	return nil, nil
}

// refreshFromParticipant updates the ResourceModel from API response.
func (r *Resource) refreshFromParticipant(data *ResourceModel, participant *shared.ParticipantResponseDto) {
	data.ParticipantID = types.Int64PointerValue(participant.ParticipantID)
	data.MemberID = types.Int64PointerValue(participant.MemberID)
	data.Email = types.StringPointerValue(participant.Email)
	data.UserName = types.StringPointerValue(participant.UserName)
	data.FirstName = types.StringPointerValue(participant.FirstName)
	data.LastName = types.StringPointerValue(participant.LastName)
	data.Role = types.StringPointerValue(participant.WspRole)
}
