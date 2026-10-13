package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringActionTriggerSourceValidator{}

// actionTriggerBlocks maps each source that takes a trigger block to the
// block's root attribute. Sources not listed (tower, github) take none.
var actionTriggerBlocks = map[string]string{
	"bucket":          "bucket",
	"cron":            "cron",
	"pipeline_status": "pipeline_status",
}

// StringActionTriggerSourceValidator checks the typed action resources'
// trigger blocks against `source`: the block matching the source must be set
// and every other trigger block must be unset. Attached to `source`.
type StringActionTriggerSourceValidator struct{}

func (v StringActionTriggerSourceValidator) Description(_ context.Context) string {
	return "Validates that only the trigger block matching source is set."
}

func (v StringActionTriggerSourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringActionTriggerSourceValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	// An unset source takes the schema default (`tower` on
	// seqera_pipeline_action), which has no trigger block.
	source := req.ConfigValue.ValueString()
	sourceDesc := fmt.Sprintf("%q", source)
	if req.ConfigValue.IsNull() {
		sourceDesc = "not set"
	}

	for blockSource, attr := range actionTriggerBlocks {
		var block types.Object
		if diags := req.Config.GetAttribute(ctx, path.Root(attr), &block); diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if block.IsUnknown() {
			continue
		}
		switch {
		case blockSource == source && block.IsNull():
			resp.Diagnostics.AddAttributeError(path.Root(attr), "Missing trigger block",
				fmt.Sprintf("An action with `source = %q` must set `%s`.", source, attr))
		case blockSource != source && !block.IsNull():
			resp.Diagnostics.AddAttributeError(path.Root(attr), "Trigger block does not match source",
				fmt.Sprintf("`%s` is only allowed when `source = %q`, but source is %s.", attr, blockSource, sourceDesc))
		}
	}
}

func ActionTriggerSourceValidator() validator.String {
	return StringActionTriggerSourceValidator{}
}
