// Package dataset_version provides the seqera_dataset_version resource.
package dataset_version

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	WorkspaceID types.Int64  `tfsdk:"workspace_id"`
	DatasetID   types.String `tfsdk:"dataset_id"`
	FilePath    types.String `tfsdk:"file_path"`
	FileHash    types.String `tfsdk:"file_hash"`
	HasHeader   types.Bool   `tfsdk:"has_header"`
	Version     types.Int64  `tfsdk:"version"`
	FileName    types.String `tfsdk:"file_name"`
	MediaType   types.String `tfsdk:"media_type"`
	URL         types.String `tfsdk:"url"`
	DateCreated types.String `tfsdk:"date_created"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset_version"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manage dataset versions in Seqera Platform.

Dataset versions represent different versions of data files uploaded to a dataset.
Each upload creates a new version. Versions can be disabled but not truly deleted.

Note: The dataset must already exist before uploading a version to it.

Import format: workspace_id/dataset_id/version (e.g., "12345/my-dataset/1")
`,
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: `Workspace numeric identifier.`,
			},
			"dataset_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: `Dataset string identifier.`,
			},
			"file_path": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: `Path to the file to upload as a new dataset version.`,
			},
			"file_hash": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `SHA256 hash of the uploaded file content. Changes trigger replacement.`,
			},
			"has_header": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
				Description: `Whether the uploaded file has a header row. Defaults to true.`,
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: `Version number of the uploaded dataset.`,
			},
			"file_name": schema.StringAttribute{
				Computed:    true,
				Description: `Name of the uploaded file.`,
			},
			"media_type": schema.StringAttribute{
				Computed:    true,
				Description: `MIME type of the uploaded file.`,
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: `URL to access the dataset version.`,
			},
			"date_created": schema.StringAttribute{
				Computed:    true,
				Description: `Timestamp when the version was created.`,
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

	// Read the file content
	filePath := data.FilePath.ValueString()
	content, err := os.ReadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read file", fmt.Sprintf("Could not read file %s: %s", filePath, err.Error()))
		return
	}

	// Compute file hash for change detection
	hash := sha256.Sum256(content)
	data.FileHash = types.StringValue(hex.EncodeToString(hash[:]))

	fileName := filepath.Base(filePath)
	hasHeader := data.HasHeader.ValueBool()
	workspaceID := data.WorkspaceID.ValueInt64()

	uploadRes, err := r.client.Datasets.UploadDatasetV2(ctx, operations.UploadDatasetV2Request{
		WorkspaceID: &workspaceID,
		DatasetID:   data.DatasetID.ValueString(),
		Header:      &hasHeader,
		MultiRequestFileSchema: shared.MultiRequestFileSchema{
			File: &shared.File{
				FileName: fileName,
				Content:  content,
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to upload dataset version", err.Error())
		return
	}
	if uploadRes.StatusCode != 200 || uploadRes.UploadDatasetVersionResponse == nil || uploadRes.UploadDatasetVersionResponse.Version == nil {
		resp.Diagnostics.AddError("Unexpected API response", common.DebugResponse(uploadRes.RawResponse))
		return
	}

	r.refreshFromVersion(&data, uploadRes.UploadDatasetVersionResponse.Version)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := data.WorkspaceID.ValueInt64()
	version, err := r.findVersion(ctx, workspaceID, data.DatasetID.ValueString(), data.Version.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read dataset version", err.Error())
		return
	}
	if version == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.refreshFromVersion(&data, version)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes require replacement, so this should never be called
	resp.Diagnostics.AddError("Update Not Supported", "Dataset version resources cannot be updated in place.")
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := data.WorkspaceID.ValueInt64()
	disableRes, err := r.client.Datasets.DisableDatasetVersion(ctx, operations.DisableDatasetVersionRequest{
		WorkspaceID: &workspaceID,
		DatasetID:   data.DatasetID.ValueString(),
		Version:     data.Version.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to disable dataset version", err.Error())
		return
	}
	if disableRes.StatusCode != 204 && disableRes.StatusCode != 404 {
		resp.Diagnostics.AddError("Failed to disable dataset version", common.DebugResponse(disableRes.RawResponse))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: workspace_id/dataset_id/version
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: workspace_id/dataset_id/version, got: %s", req.ID),
		)
		return
	}

	workspaceID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid workspace_id",
			fmt.Sprintf("workspace_id must be a number, got: %s", parts[0]),
		)
		return
	}

	version, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid version",
			fmt.Sprintf("version must be a number, got: %s", parts[2]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dataset_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version)...)
	// file_path and file_hash cannot be recovered from import - user must set file_path in config
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("file_path"), types.StringNull())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("file_hash"), types.StringNull())...)
}

// findVersion searches for a dataset version by version number.
func (r *Resource) findVersion(ctx context.Context, workspaceID int64, datasetID string, versionNum int64) (*shared.DatasetVersionDto, error) {
	listRes, err := r.client.Datasets.ListDatasetVersionsV2(ctx, operations.ListDatasetVersionsV2Request{
		WorkspaceID: &workspaceID,
		DatasetID:   datasetID,
	})
	if err != nil {
		return nil, err
	}
	if listRes.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d listing dataset versions", listRes.StatusCode)
	}
	if listRes.ListDatasetVersionsResponse == nil {
		return nil, fmt.Errorf("empty response listing dataset versions")
	}

	for i := range listRes.ListDatasetVersionsResponse.Versions {
		v := &listRes.ListDatasetVersionsResponse.Versions[i]
		if v.Version != nil && *v.Version == versionNum {
			// Skip disabled versions
			if v.Disabled != nil && *v.Disabled {
				return nil, nil
			}
			return v, nil
		}
	}
	return nil, nil
}

// refreshFromVersion updates the ResourceModel from API response.
func (r *Resource) refreshFromVersion(data *ResourceModel, version *shared.DatasetVersionDto) {
	data.Version = types.Int64PointerValue(version.Version)
	data.FileName = types.StringPointerValue(version.FileName)
	data.MediaType = types.StringPointerValue(version.MediaType)
	data.URL = types.StringPointerValue(version.URL)
	if version.HasHeader != nil {
		data.HasHeader = types.BoolPointerValue(version.HasHeader)
	}
	if version.DateCreated != nil {
		data.DateCreated = types.StringValue(version.DateCreated.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		data.DateCreated = types.StringNull()
	}
}
