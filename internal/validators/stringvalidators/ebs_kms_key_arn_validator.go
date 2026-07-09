package stringvalidators

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringEbsKmsKeyArnValidator{}

// kmsKeyArnPattern mirrors KMS_KEY_ARN_PATTERN in the platform backend
// (AwsCloudPlatformProvider): a well-formed KMS key ARN, allowing the aws /
// aws-cn / aws-us-gov partitions, a standard UUID key id, or a multi-region
// (mrk-) key id.
var kmsKeyArnPattern = regexp.MustCompile(`^arn:aws[a-z0-9-]*:kms:[a-z0-9-]+:\d{12}:key/(mrk-[0-9a-f]{32}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)

// StringEbsKmsKeyArnValidator validates that ebs_kms_key_id, when set, is a
// well-formed KMS key ARN. It matches the backend's save-time check so a
// malformed ARN is caught at plan time rather than at apply.
type StringEbsKmsKeyArnValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringEbsKmsKeyArnValidator) Description(_ context.Context) string {
	return "ebs_kms_key_id must be a well-formed KMS key ARN"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringEbsKmsKeyArnValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v StringEbsKmsKeyArnValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// An empty/blank value is always valid (falls back to the account default key).
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}

	if !kmsKeyArnPattern.MatchString(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid KMS Key ARN",
			"ebs_kms_key_id must be a well-formed KMS key ARN, e.g. "+
				"arn:aws:kms:<region>:<account-id>:key/<key-id>. Leave it empty to use the account/region default EBS encryption key.",
		)
	}
}

// EbsKmsKeyArnValidator returns a validator that checks ebs_kms_key_id is a
// well-formed KMS key ARN.
func EbsKmsKeyArnValidator() validator.String {
	return StringEbsKmsKeyArnValidator{}
}
