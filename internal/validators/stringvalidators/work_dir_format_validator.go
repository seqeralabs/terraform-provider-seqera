package stringvalidators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringWorkDirFormatValidatorValidator{}

type StringWorkDirFormatValidatorValidator struct{}

func (v StringWorkDirFormatValidatorValidator) Description(_ context.Context) string {
	return "validates that work_dir starts with a valid cloud storage prefix (s3://, gs://, az://) or is an absolute local path (/), and does not end with a trailing slash"
}

func (v StringWorkDirFormatValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringWorkDirFormatValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	workDir := req.ConfigValue.ValueString()
	if workDir == "" {
		return
	}

	// Validate that work_dir starts with a valid prefix
	validPrefixes := []string{"s3://", "gs://", "az://", "/"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(workDir, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Working Directory",
			fmt.Sprintf(
				"work_dir must start with a valid cloud storage prefix (s3://, gs://, az://) or be an absolute local path (/)."+
					" Got: %q", workDir,
			),
		)
		return
	}

	// Check for trailing slash — the API strips trailing slashes at launch time,
	// which causes plan diffs if the stored value has one.
	// Only check paths longer than "/" itself.
	if len(workDir) > 1 && strings.HasSuffix(workDir, "/") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Trailing Slash in Working Directory",
			fmt.Sprintf(
				"work_dir should not end with a trailing slash. The Seqera API strips trailing slashes,"+
					" which would cause unexpected plan diffs. Please remove the trailing slash: %q → %q",
				workDir, strings.TrimRight(workDir, "/"),
			),
		)
		return
	}

	// For cloud storage URIs, validate that there's a bucket/container name after the prefix
	cloudPrefixes := map[string]string{
		"s3://": "S3 bucket",
		"gs://": "GCS bucket",
		"az://": "Azure container",
	}
	for prefix, name := range cloudPrefixes {
		if strings.HasPrefix(workDir, prefix) {
			remainder := workDir[len(prefix):]
			if remainder == "" {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Missing Bucket/Container Name",
					fmt.Sprintf(
						"work_dir with prefix %q must include a %s name (e.g., %sexample-bucket/work). Got: %q",
						prefix, name, prefix, workDir,
					),
				)
				return
			}
			break
		}
	}
}

func WorkDirFormatValidator() validator.String {
	return StringWorkDirFormatValidatorValidator{}
}
