package int32validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	maxSpotAttemptsMin = 1
	maxSpotAttemptsMax = 10
)

var _ validator.Int32 = Int32MaxSpotAttemptsValidator{}

// Int32MaxSpotAttemptsValidator enforces the two constraints the API documents
// for `max_spot_attempts` but does not express in the schema:
//
//  1. the value must fall in the inclusive range 1-10;
//  2. it is only honoured when the provisioning model requests Spot capacity.
//
// The provisioning model defaults to `spotFirst`, so an unset
// `provisioning_model` still uses Spot — only an explicit `ondemand` makes the
// setting a silent no-op.
type Int32MaxSpotAttemptsValidator struct{}

// Description describes the validation in plain text formatting.
func (v Int32MaxSpotAttemptsValidator) Description(_ context.Context) string {
	return fmt.Sprintf(
		"value must be between %d and %d, and may only be set when provisioning_model requests Spot capacity",
		maxSpotAttemptsMin, maxSpotAttemptsMax,
	)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v Int32MaxSpotAttemptsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateInt32 performs the validation.
func (v Int32MaxSpotAttemptsValidator) ValidateInt32(ctx context.Context, req validator.Int32Request, resp *validator.Int32Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attempts := req.ConfigValue.ValueInt32()
	if attempts < maxSpotAttemptsMin || attempts > maxSpotAttemptsMax {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid max_spot_attempts",
			fmt.Sprintf(
				"`max_spot_attempts` must be between %d and %d (inclusive), got %d. "+
					"%d means a single Spot attempt with no retry.",
				maxSpotAttemptsMin, maxSpotAttemptsMax, attempts, maxSpotAttemptsMin,
			),
		)
		return
	}

	var provisioningModel types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("provisioning_model"), &provisioningModel)...)
	if resp.Diagnostics.HasError() || provisioningModel.IsUnknown() {
		return
	}

	// Unset means the API default (`spotFirst`), which does use Spot.
	if provisioningModel.IsNull() {
		return
	}

	if provisioningModel.ValueString() == "ondemand" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Unused max_spot_attempts",
			"`max_spot_attempts` only applies when Spot capacity is requested, but "+
				"`provisioning_model` is `ondemand`. Remove `max_spot_attempts`, or set "+
				"`provisioning_model` to `spot` or `spotFirst`.",
		)
	}
}

// MaxSpotAttemptsValidator returns a validator which ensures max_spot_attempts
// is within the API's allowed range and is not set alongside an on-demand-only
// provisioning model.
func MaxSpotAttemptsValidator() validator.Int32 {
	return Int32MaxSpotAttemptsValidator{}
}
