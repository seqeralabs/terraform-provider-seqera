// Package service_account_secret provides the seqera_service_account_secret
// resource, which issues a client secret for a service account.
//
// The resource is hand-written because the API has no endpoint that reads one
// secret by ID: Read lists the account's secrets and matches on the ID kept in
// state. A Speakeasy-generated read would take the first list item, which is
// the older secret while two are live during a rotation.
//
// The schema mirrors what an overlay-generated resource could express, so the
// resource can move to Speakeasy without a state change if Platform adds
// GET /orgs/{orgId}/service-accounts/{saId}/secrets/{secretId}. The exception
// is client_id, which the issue response does not carry: Create reads it from
// the service account, because Platform creates the OAuth client lazily on the
// first secret, so seqera_service_account.client_credentials.client_id is
// still null in the same apply.
package service_account_secret

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	OrgID            types.Int64  `tfsdk:"org_id"`
	ServiceAccountID types.Int64  `tfsdk:"service_account_id"`
	ExpiryDays       types.Int64  `tfsdk:"expiry_days"`
	ID               types.Int64  `tfsdk:"id"`
	ClientID         types.String `tfsdk:"client_id"`
	Secret           types.String `tfsdk:"secret"`
	Hint             types.String `tfsdk:"hint"`
	Status           types.String `tfsdk:"status"`
	CreatedAt        types.String `tfsdk:"created_at"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_secret"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Issue a client secret for a service account.

The service account exchanges its client ID (` + "`client_id`" + `) and this secret for an access token at ` + "`POST /token`" + `. The secret value is returned only when it is issued,
so it is kept in the Terraform state as a sensitive attribute: protect the state accordingly.

Every argument forces a new secret. Destroying the resource revokes the secret; access tokens
already issued with it stay valid until they expire (1 hour by default).

A service account holds at most 2 live secrets. To rotate without downtime, manage one secret per
service account with ` + "`create_before_destroy`" + ` and trigger replacement with ` + "`time_rotating`" + `:
the new secret is issued before the old one is revoked. Managing more than one secret per service
account, or issuing secrets outside Terraform, can fail an apply with a conflict.

A secret that expires or is revoked outside Terraform is removed from the state, so the next plan
issues a new one.

Import is not supported: the secret value cannot be retrieved after it is issued.
`,
		Attributes: map[string]schema.Attribute{
			"org_id": schema.Int64Attribute{
				Required:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Description:   `Organization numeric identifier. Changing it issues a new secret.`,
			},
			"service_account_id": schema.Int64Attribute{
				Required:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Description:   "Service account numeric identifier (`seqera_service_account.x.id`). Changing it issues a new secret.",
			},
			"expiry_days": schema.Int64Attribute{
				Required:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{int64validator.OneOf(30, 90, 180, 365)},
				Description:   `Lifetime of the secret in days: 30, 90, 180 or 365. Changing it issues a new secret.`,
			},
			"id": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Description:   `Secret numeric identifier.`,
			},
			"client_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description: "OAuth client ID of the service account, to use with `secret`. Platform creates the client " +
					"when the first secret is issued, so prefer this attribute over " +
					"`seqera_service_account.client_credentials.client_id`, which stays null until the next refresh.",
			},
			"secret": schema.StringAttribute{
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   `Secret value. Returned by the API only when the secret is issued.`,
			},
			"hint": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   `Masked secret showing its last 4 characters, as shown in the Platform UI.`,
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the secret: `ACTIVE`, or `EXPIRING` when it is close to its expiry date.",
			},
			"created_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   `When the secret was issued (RFC 3339).`,
			},
			"expires_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   `When the secret expires (RFC 3339). From then on it can no longer be exchanged for an access token.`,
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

	orgID, saID := data.OrgID.ValueInt64(), data.ServiceAccountID.ValueInt64()
	res, err := r.client.ServiceAccounts.IssueServiceAccountSecret(ctx, operations.IssueServiceAccountSecretRequest{
		OrgID: orgID,
		SaID:  saID,
		IssueServiceAccountSecretRequest: shared.IssueServiceAccountSecretRequest{
			ExpiryDays: int(data.ExpiryDays.ValueInt64()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to issue service account secret", err.Error())
		return
	}
	switch {
	case res.StatusCode == http.StatusCreated && res.IssueServiceAccountSecretResponse != nil:
	case res.StatusCode == http.StatusConflict:
		resp.Diagnostics.AddError("Cannot issue service account secret",
			fmt.Sprintf("Service account %d already has 2 live secrets, or it is disabled. "+
				"Revoke a secret or enable the service account, then apply again. "+
				"To rotate, manage one secret per service account with create_before_destroy.\n\n%s",
				saID, common.DebugResponse(res.RawResponse)))
		return
	default:
		common.AddUnexpectedStatus(&resp.Diagnostics, "issuing service account secret", res.RawResponse)
		return
	}

	issued := res.IssueServiceAccountSecretResponse
	data.ID = types.Int64Value(issued.SecretID)
	data.Secret = types.StringValue(issued.Secret)
	data.Hint = types.StringValue(issued.Hint)
	data.ExpiresAt = types.StringValue(formatTime(issued.ExpiresAt))

	// The issue response has no status, creation date or client ID: read
	// them back. The secret is already issued, so a failure here keeps it in
	// state (with the value, which cannot be fetched again) and only warns;
	// Read fills in the client ID later.
	secret, err := find(ctx, r.client, orgID, saID, issued.SecretID)
	if err == nil && secret != nil {
		refresh(&data, secret)
	} else {
		data.Status = types.StringValue(string(shared.ClientSecretStatusActive))
		data.CreatedAt = types.StringNull()
		detail := "the secret was not listed"
		if err != nil {
			detail = err.Error()
		}
		resp.Diagnostics.AddWarning("Could not read back the issued secret",
			"The secret was issued and stored in the state, but its status and creation date could not be read: "+detail)
	}
	data.ClientID = types.StringNull()
	if id, err := clientID(ctx, r.client, orgID, saID); err != nil {
		resp.Diagnostics.AddWarning("Could not read the service account client ID",
			"The secret was issued and stored in the state, but client_id could not be read. "+
				"It is filled in on the next refresh: "+err.Error())
	} else {
		data.ClientID = types.StringPointerValue(id)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := find(ctx, r.client, data.OrgID.ValueInt64(), data.ServiceAccountID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read service account secret", err.Error())
		return
	}
	if !live(secret) {
		resp.State.RemoveResource(ctx)
		return
	}

	refresh(&data, secret)
	// A service account keeps one OAuth client for life, so the client ID is
	// only fetched when Create could not read it.
	if data.ClientID.IsNull() || data.ClientID.IsUnknown() {
		id, err := clientID(ctx, r.client, data.OrgID.ValueInt64(), data.ServiceAccountID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read service account client ID", err.Error())
			return
		}
		data.ClientID = types.StringPointerValue(id)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Every argument requires replacement, so Terraform never plans an update.
	resp.Diagnostics.AddError("Update not supported", "Service account secrets cannot be updated in place.")
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.ServiceAccounts.RevokeServiceAccountSecret(ctx, operations.RevokeServiceAccountSecretRequest{
		OrgID:    data.OrgID.ValueInt64(),
		SaID:     data.ServiceAccountID.ValueInt64(),
		SecretID: data.ID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to revoke service account secret", err.Error())
		return
	}
	switch res.StatusCode {
	// 404: the secret or its service account is gone. 409: the secret has
	// already expired. Either way it can no longer be used.
	case http.StatusOK, http.StatusNotFound, http.StatusConflict:
	default:
		common.AddUnexpectedStatus(&resp.Diagnostics, "revoking service account secret", res.RawResponse)
	}
}

// find returns the secret with the given ID, or nil when the service account
// does not list it. The list holds every live secret (at most 2) plus secrets
// that ended in the last 30 days, so a missing ID means the secret has ended.
// A 404 means the service account is gone, which also ends its secrets.
func find(ctx context.Context, client *sdk.Seqera, orgID, saID, secretID int64) (*shared.ClientSecretDto, error) {
	res, err := client.ServiceAccounts.ListServiceAccountSecrets(ctx, operations.ListServiceAccountSecretsRequest{
		OrgID: orgID,
		SaID:  saID,
	})
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK || res.ListServiceAccountSecretsResponse == nil {
		return nil, common.UnexpectedStatusErr("listing service account secrets", res.RawResponse)
	}
	for i := range res.ListServiceAccountSecretsResponse.Secrets {
		if s := &res.ListServiceAccountSecretsResponse.Secrets[i]; s.ID == secretID {
			return s, nil
		}
	}
	return nil, nil
}

// clientID returns the OAuth client ID of the service account, or nil when it
// has no client yet.
func clientID(ctx context.Context, client *sdk.Seqera, orgID, saID int64) (*string, error) {
	res, err := client.ServiceAccounts.DescribeServiceAccount(ctx, operations.DescribeServiceAccountRequest{
		OrgID: orgID,
		SaID:  saID,
	})
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK || res.DescribeServiceAccountResponse == nil {
		return nil, common.UnexpectedStatusErr("describing service account", res.RawResponse)
	}
	sa := res.DescribeServiceAccountResponse.ServiceAccount
	if sa == nil || sa.ClientCredentials == nil {
		return nil, nil
	}
	return sa.ClientCredentials.ClientID, nil
}

// live reports whether the secret can still be exchanged for an access token.
func live(s *shared.ClientSecretDto) bool {
	if s == nil {
		return false
	}
	return s.Status == shared.ClientSecretStatusActive || s.Status == shared.ClientSecretStatusExpiring
}

func refresh(data *ResourceModel, s *shared.ClientSecretDto) {
	data.ID = types.Int64Value(s.ID)
	data.Hint = types.StringValue(s.Hint)
	data.Status = types.StringValue(string(s.Status))
	data.CreatedAt = types.StringValue(formatTime(s.CreatedAt))
	data.ExpiresAt = types.StringValue(formatTime(s.ExpiresAt))
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
