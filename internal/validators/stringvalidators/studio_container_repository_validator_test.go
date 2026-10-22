package stringvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStudioContainerRepositoryValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     types.String
		wantError bool
	}{
		"null":                {types.StringNull(), false},
		"unknown":             {types.StringUnknown(), false},
		"bare registry path":  {types.StringValue("cr.seqera.io/my-org/studios"), false},
		"registry with port":  {types.StringValue("registry.local:5000/studios"), false},
		"https scheme":        {types.StringValue("https://cr.seqera.io/my-org/studios"), true},
		"http scheme":         {types.StringValue("http://cr.seqera.io/my-org/studios"), true},
		"uppercase scheme":    {types.StringValue("HTTPS://cr.seqera.io/studios"), true},
		"host starting http":  {types.StringValue("httpbin.example.com/studios"), false},
		"scheme not at start": {types.StringValue("cr.seqera.io/https://x"), false},
		"empty":               {types.StringValue(""), false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{Path: path.Root("container_repository"), ConfigValue: tc.value}
			resp := &validator.StringResponse{}
			StudioContainerRepositoryValidator().ValidateString(context.Background(), req, resp)

			if got := resp.Diagnostics.HasError(); got != tc.wantError {
				t.Fatalf("HasError() = %v, want %v (diags: %v)", got, tc.wantError, resp.Diagnostics)
			}
		})
	}
}
