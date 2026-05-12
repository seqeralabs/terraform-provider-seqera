package stringvalidators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringWorkDirS3ValidatorValidator{}

type StringWorkDirS3ValidatorValidator struct{}

func (v StringWorkDirS3ValidatorValidator) Description(_ context.Context) string {
	return "validates that work_dir is an s3:// URI with a bucket name and no trailing slash"
}

func (v StringWorkDirS3ValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringWorkDirS3ValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	workDir := req.ConfigValue.ValueString()
	if workDir == "" {
		return
	}

	if !strings.HasPrefix(workDir, "s3://") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid AWS Working Directory",
			fmt.Sprintf(
				"work_dir must be an s3:// URI for AWS compute environments (e.g., s3://my-bucket/work). Got: %q",
				workDir,
			),
		)
		return
	}

	if workDir == "s3://" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing S3 Bucket Name",
			fmt.Sprintf(
				"work_dir must include an S3 bucket name after the s3:// prefix (e.g., s3://my-bucket/work). Got: %q",
				workDir,
			),
		)
		return
	}

	// The Seqera API strips trailing slashes; mismatched stored values cause plan diffs.
	if strings.HasSuffix(workDir, "/") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Trailing Slash in Working Directory",
			fmt.Sprintf(
				"work_dir should not end with a trailing slash. The Seqera API strips trailing slashes, which would cause unexpected plan diffs. Please remove the trailing slash: %q → %q",
				workDir, strings.TrimRight(workDir, "/"),
			),
		)
		return
	}
}

func WorkDirS3Validator() validator.String {
	return StringWorkDirS3ValidatorValidator{}
}
