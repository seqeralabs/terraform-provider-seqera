package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringBitbucketTokenValidatorValidator{}

type StringBitbucketTokenValidatorValidator struct{}

func (v StringBitbucketTokenValidatorValidator) Description(_ context.Context) string {
	return "validates that `token` and `password` are not both set on a Bitbucket credential"
}

func (v StringBitbucketTokenValidatorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v StringBitbucketTokenValidatorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateBitbucketKeysSibling(ctx, req, resp, "password")
}

func BitbucketTokenValidator() validator.String {
	return StringBitbucketTokenValidatorValidator{}
}
