package stringvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Platform Cloud default; Enterprise installs can override.
const runScriptSizeSoftLimit = 1024

var runScriptSizeDescription = fmt.Sprintf(
	"warns when a pre/post-run script exceeds %d bytes (Platform Cloud default)",
	runScriptSizeSoftLimit,
)

var _ validator.String = StringRunScriptSizeValidatorValidator{}

type StringRunScriptSizeValidatorValidator struct{}

func (v StringRunScriptSizeValidatorValidator) Description(_ context.Context) string {
	return runScriptSizeDescription
}

func (v StringRunScriptSizeValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringRunScriptSizeValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	size := len(req.ConfigValue.ValueString())
	if size <= runScriptSizeSoftLimit {
		return
	}

	resp.Diagnostics.AddAttributeWarning(
		req.Path,
		"script may be rejected by Platform Cloud for size",
		fmt.Sprintf(
			"Script is %d bytes; Seqera Platform Cloud rejects pre/post-run scripts above %d bytes "+
				"(Enterprise installs may raise this). Consider hosting larger logic in object storage "+
				"(S3, GCS, Azure Blob) and downloading from a short script.",
			size, runScriptSizeSoftLimit,
		),
	)
}

func RunScriptSizeValidator() validator.String {
	return StringRunScriptSizeValidatorValidator{}
}
