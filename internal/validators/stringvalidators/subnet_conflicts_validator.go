package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringSubnetConflictsValidator{}

// StringSubnetConflictsValidator validates that the deprecated single-value
// subnet_id is not combined with the multi-value subnet_ids. It mirrors the
// tower-cli guard (AwsCloudPlatform: "Options --subnet-id and --subnet-ids are
// mutually exclusive; use --subnet-ids").
type StringSubnetConflictsValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringSubnetConflictsValidator) Description(_ context.Context) string {
	return "subnet_id and subnet_ids are mutually exclusive"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringSubnetConflictsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v StringSubnetConflictsValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Nothing to check if the deprecated subnet_id is not configured.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}

	var subnetIds types.List
	siblingPath := req.Path.ParentPath().AtName("subnet_ids")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &subnetIds)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase (for_each, count, etc.).
	if subnetIds.IsUnknown() {
		return
	}

	if !subnetIds.IsNull() && len(subnetIds.Elements()) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting Subnet Configuration",
			"subnet_id and subnet_ids are mutually exclusive. subnet_id is deprecated; use subnet_ids instead.",
		)
	}
}

// SubnetConflictsValidator returns a validator ensuring subnet_id is not set
// together with subnet_ids.
func SubnetConflictsValidator() validator.String {
	return StringSubnetConflictsValidator{}
}
