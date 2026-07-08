package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = StringEbsKmsKeyEncryptionValidator{}

// StringEbsKmsKeyEncryptionValidator validates that an EBS KMS key ARN is only
// supplied when boot-disk encryption is enabled. It mirrors the tower-cli guard
// (AwsCloudPlatform: "EBS KMS key requires EBS encryption to be enabled").
type StringEbsKmsKeyEncryptionValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringEbsKmsKeyEncryptionValidator) Description(_ context.Context) string {
	return "ebs_kms_key_id can only be set when ebs_encrypted is true"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringEbsKmsKeyEncryptionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v StringEbsKmsKeyEncryptionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Nothing to check if no KMS key is configured.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}

	var ebsEncrypted types.Bool
	siblingPath := req.Path.ParentPath().AtName("ebs_encrypted")
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &ebsEncrypted)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Allow unknown values during plan phase (for_each, count, etc.).
	if ebsEncrypted.IsUnknown() {
		return
	}

	if ebsEncrypted.IsNull() || !ebsEncrypted.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid EBS Encryption Configuration",
			"ebs_kms_key_id requires EBS boot-disk encryption to be enabled. Set ebs_encrypted = true, or remove ebs_kms_key_id to use the account/region default EBS encryption key.",
		)
	}
}

// EbsKmsKeyEncryptionValidator returns a validator ensuring ebs_kms_key_id is
// only set together with ebs_encrypted = true.
func EbsKmsKeyEncryptionValidator() validator.String {
	return StringEbsKmsKeyEncryptionValidator{}
}
