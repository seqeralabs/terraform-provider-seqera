// Custom data source for listing and filtering credentials by name

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tfTypes "github.com/seqeralabs/terraform-provider-seqera/internal/provider/types"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &CredentialsDataSource{}
	_ datasource.DataSourceWithConfigure = &CredentialsDataSource{}
)

func NewCredentialsDataSource() datasource.DataSource {
	return &CredentialsDataSource{}
}

// CredentialsDataSource is the data source implementation for listing credentials.
type CredentialsDataSource struct {
	// Provider configured SDK client.
	client *sdk.Seqera
}

// CredentialsDataSourceModel describes the data model.
type CredentialsDataSourceModel struct {
	// Optional filters
	WorkspaceID types.Int64  `tfsdk:"workspace_id"`
	PlatformID  types.String `tfsdk:"platform_id"`
	Name        types.String `tfsdk:"name"`

	// Results
	Credentials []CredentialModel `tfsdk:"credentials"`
	ID          types.String      `tfsdk:"id"`
}

// CredentialModel represents a single credential in the list
type CredentialModel struct {
	BaseURL       types.String               `tfsdk:"base_url"`
	Category      types.String               `tfsdk:"category"`
	CredentialsID types.String               `tfsdk:"credentials_id"`
	DateCreated   types.String               `tfsdk:"date_created"`
	Deleted       types.Bool                 `tfsdk:"deleted"`
	Description   types.String               `tfsdk:"description"`
	Keys          tfTypes.SecurityKeysOutput `tfsdk:"keys"`
	LastUpdated   types.String               `tfsdk:"last_updated"`
	LastUsed      types.String               `tfsdk:"last_used"`
	Name          types.String               `tfsdk:"name"`
	ProviderType  types.String               `tfsdk:"provider_type"`
}

// Metadata returns the data source type name.
func (r *CredentialsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credentials"
}

// Schema defines the schema for the data source.
func (r *CredentialsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `List and filter workspace credentials in Seqera platform.

This data source allows you to retrieve credentials by name or list all credentials
in a workspace. Use this when you need to reference existing credentials by name
rather than by their ID.

## Example Usage

` + "```terraform" + `
# Fetch all credentials in a workspace
data "seqera_credentials" "all" {
  workspace_id = var.seqera_workspace_id
}

# Fetch specific credential by name
data "seqera_credentials" "by_name" {
  workspace_id = var.seqera_workspace_id
  name         = "my-aws-credentials"
}

# Use the credential ID in a compute environment
resource "seqera_compute_env" "example" {
  name           = "my-compute-env"
  workspace_id   = var.seqera_workspace_id
  credentials_id = data.seqera_credentials.by_name.credentials[0].credentials_id
  # ... other configuration
}
` + "```",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID for the data source",
			},
			"workspace_id": schema.Int64Attribute{
				Optional:    true,
				Description: `Workspace numeric identifier`,
			},
			"platform_id": schema.StringAttribute{
				Optional:    true,
				Description: `Platform string identifier to filter credentials`,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: `Filter credentials by name (exact match)`,
			},
			"credentials": schema.ListNestedAttribute{
				Computed:    true,
				Description: `List of credentials matching the filter criteria`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"base_url": schema.StringAttribute{
							Computed: true,
						},
						"category": schema.StringAttribute{
							Computed: true,
						},
						"credentials_id": schema.StringAttribute{
							Computed:    true,
							Description: `Credentials string identifier`,
						},
						"date_created": schema.StringAttribute{
							Computed:    true,
							Description: `Timestamp when the credential was created`,
						},
						"deleted": schema.BoolAttribute{
							Computed:    true,
							Description: `Flag indicating if the credential has been soft-deleted`,
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: `Optional description explaining the purpose of the credential`,
						},
						"keys": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"aws": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"access_key": schema.StringAttribute{
											Computed: true,
										},
										"assume_role_arn": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"azure": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"batch_name": schema.StringAttribute{
											Computed: true,
										},
										"storage_name": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"azure_entra": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"batch_name": schema.StringAttribute{
											Computed: true,
										},
										"client_id": schema.StringAttribute{
											Computed: true,
										},
										"storage_name": schema.StringAttribute{
											Computed: true,
										},
										"tenant_id": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"azurerepos": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"username": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"bitbucket": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"username": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"codecommit": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"username": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"container_reg": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"registry": schema.StringAttribute{
											Computed: true,
										},
										"user_name": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"gitea": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"username": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"github": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"username": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"gitlab": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"username": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"google": schema.SingleNestedAttribute{
									Computed: true,
								},
								"k8s": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"certificate": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"seqeracompute": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"access_key": schema.StringAttribute{
											Computed: true,
										},
										"assume_role_arn": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"ssh": schema.SingleNestedAttribute{
									Computed: true,
								},
								"tw_agent": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"connection_id": schema.StringAttribute{
											Computed: true,
										},
										"shared": schema.BoolAttribute{
											Computed: true,
										},
										"work_dir": schema.StringAttribute{
											Computed: true,
										},
									},
								},
							},
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
						"last_used": schema.StringAttribute{
							Computed:    true,
							Description: `Timestamp when the credential was last used`,
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: `Display name for the credential (max 100 characters)`,
						},
						"provider_type": schema.StringAttribute{
							Computed:    true,
							Description: `Cloud or service provider type (e.g., aws, azure, gcp)`,
						},
					},
				},
			},
		},
	}
}

func (r *CredentialsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.Seqera)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *sdk.Seqera, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *CredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *CredentialsDataSourceModel
	var item types.Object

	resp.Diagnostics.Append(req.Config.Get(ctx, &item)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(item.As(ctx, &data, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request for listing credentials
	request := operations.ListCredentialsRequest{}

	if !data.WorkspaceID.IsNull() {
		workspaceId := data.WorkspaceID.ValueInt64()
		request.WorkspaceID = &workspaceId
	}

	if !data.PlatformID.IsNull() {
		platformId := data.PlatformID.ValueString()
		request.PlatformID = &platformId
	}

	// Call the ListCredentials API
	res, err := r.client.Credentials.Credentials(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke API", err.Error())
		if res != nil && res.RawResponse != nil {
			resp.Diagnostics.AddError("unexpected http request/response", debugResponse(res.RawResponse))
		}
		return
	}
	if res == nil {
		resp.Diagnostics.AddError("unexpected response from API", fmt.Sprintf("%v", res))
		return
	}
	if res.StatusCode != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from API. Got an unexpected response code %v", res.StatusCode), debugResponse(res.RawResponse))
		return
	}
	if res.ListCredentialsResponse == nil {
		resp.Diagnostics.AddError("unexpected response from API. Got an unexpected response body", debugResponse(res.RawResponse))
		return
	}

	// Filter credentials by name if specified
	filteredCredentials := res.ListCredentialsResponse.GetCredentials()
	if !data.Name.IsNull() {
		targetName := data.Name.ValueString()
		var matchingCredentials []shared.CredentialsOutput

		for _, cred := range filteredCredentials {
			if cred.GetName() == targetName {
				matchingCredentials = append(matchingCredentials, cred)
			}
		}
		filteredCredentials = matchingCredentials
	}

	// Convert the filtered credentials to the expected model format
	var credentialModels []CredentialModel
	for _, cred := range filteredCredentials {
		credModel := CredentialModel{}

		// Convert each field
		if cred.GetCredentialsID() != nil {
			credModel.CredentialsID = types.StringValue(*cred.GetCredentialsID())
		} else {
			credModel.CredentialsID = types.StringNull()
		}

		credModel.Name = types.StringValue(cred.GetName())

		if cred.GetDescription() != nil {
			credModel.Description = types.StringValue(*cred.GetDescription())
		} else {
			credModel.Description = types.StringNull()
		}

		credModel.ProviderType = types.StringValue(string(cred.GetProviderType()))

		if cred.GetBaseURL() != nil {
			credModel.BaseURL = types.StringValue(*cred.GetBaseURL())
		} else {
			credModel.BaseURL = types.StringNull()
		}

		if cred.GetCategory() != nil {
			credModel.Category = types.StringValue(*cred.GetCategory())
		} else {
			credModel.Category = types.StringNull()
		}

		if cred.GetDeleted() != nil {
			credModel.Deleted = types.BoolValue(*cred.GetDeleted())
		} else {
			credModel.Deleted = types.BoolNull()
		}

		if cred.GetDateCreated() != nil {
			credModel.DateCreated = types.StringValue(cred.GetDateCreated().String())
		} else {
			credModel.DateCreated = types.StringNull()
		}

		if cred.GetLastUpdated() != nil {
			credModel.LastUpdated = types.StringValue(cred.GetLastUpdated().String())
		} else {
			credModel.LastUpdated = types.StringNull()
		}

		if cred.GetLastUsed() != nil {
			credModel.LastUsed = types.StringValue(cred.GetLastUsed().String())
		} else {
			credModel.LastUsed = types.StringNull()
		}

		// Set keys as null for now - this is a complex nested structure
		// Users who need detailed key information should use the individual credential data source
		credModel.Keys = tfTypes.SecurityKeysOutput{}

		credentialModels = append(credentialModels, credModel)
	}

	data.Credentials = credentialModels

	// Set ID for the data source
	if !data.Name.IsNull() {
		data.ID = types.StringValue(fmt.Sprintf("credentials-name-%s", data.Name.ValueString()))
	} else if !data.WorkspaceID.IsNull() {
		data.ID = types.StringValue(fmt.Sprintf("credentials-workspace-%d", data.WorkspaceID.ValueInt64()))
	} else {
		data.ID = types.StringValue("credentials-all")
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
