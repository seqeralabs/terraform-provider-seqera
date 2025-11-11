package objectvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tfTypes "github.com/seqeralabs/terraform-provider-seqera/internal/provider/types"
)

var _ validator.Object = ObjectConfigEnvVariableValidator{}

type ObjectConfigEnvVariableValidator struct{}

// Description describes the validation in plain text formatting.
func (v ObjectConfigEnvVariableValidator) Description(_ context.Context) string {
	return "Validates that at least one of 'head' or 'compute' must be set to true for environment variables. Both can be true to target both head and compute nodes."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v ObjectConfigEnvVariableValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v ObjectConfigEnvVariableValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Skip validation if the value is null or unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Extract the ConfigEnvVariable from the request
	var envVar tfTypes.ConfigEnvVariable
	diags := req.ConfigValue.As(ctx, &envVar, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Rule: At least one of head or compute must be true
	headIsTrue := envVar.Head.ValueBool()
	computeIsTrue := envVar.Compute.ValueBool()

	if !headIsTrue && !computeIsTrue {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Environment Variable Configuration",
			"At least one of 'head' or 'compute' must be set to true. Environment variables must target the head node, compute nodes, or both.",
		)
	}
}

func ConfigEnvVariableValidator() validator.Object {
	return ObjectConfigEnvVariableValidator{}
}
