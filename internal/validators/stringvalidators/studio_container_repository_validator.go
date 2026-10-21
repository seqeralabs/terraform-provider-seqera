package stringvalidators

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringStudioContainerRepositoryValidator{}

// StringStudioContainerRepositoryValidator rejects a container repository that
// starts with an http:// or https:// scheme. Platform requires a bare registry
// path such as "cr.seqera.io/my-org/studios" and answers anything else with a
// 400 at apply time. Go's RE2 has no negative lookahead, so a regex pattern in
// the OpenAPI spec cannot express this.
type StringStudioContainerRepositoryValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringStudioContainerRepositoryValidator) Description(_ context.Context) string {
	return "container repository must not start with http:// or https://"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringStudioContainerRepositoryValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v StringStudioContainerRepositoryValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	lower := strings.ToLower(req.ConfigValue.ValueString())
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Container Repository",
			"The container repository must not include the http:// or https:// protocol. Use the registry path only, for example 'cr.seqera.io/my-org/studios'.",
		)
	}
}

// StudioContainerRepositoryValidator returns a validator that rejects a
// container repository starting with http:// or https://.
func StudioContainerRepositoryValidator() validator.String {
	return StringStudioContainerRepositoryValidator{}
}
