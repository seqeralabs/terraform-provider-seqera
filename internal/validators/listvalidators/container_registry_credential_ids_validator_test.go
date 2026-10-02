package listvalidators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestContainerRegistryCredentialIdsValidator(t *testing.T) {
	tests := []struct {
		name        string
		value       types.List
		expectError bool
	}{
		{name: "null list", value: types.ListNull(types.StringType)},
		{name: "unknown list", value: types.ListUnknown(types.StringType)},
		{
			name: "Seqera credential IDs",
			value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("0"),
				types.StringValue(strings.Repeat("x", 22)),
			}),
		},
		{
			name: "unknown element",
			value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringUnknown(),
			}),
		},
		{
			name: "null element",
			value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringNull(),
			}),
			expectError: true,
		},
		{
			name: "empty value",
			value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(""),
			}),
			expectError: true,
		},
		{
			name: "value is too long",
			value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(strings.Repeat("x", 23)),
			}),
			expectError: true,
		},
		{
			name: "value contains non-Base62 characters",
			value: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("not/a/credential/id"),
			}),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &validator.ListResponse{}
			ContainerRegistryCredentialIdsValidator().ValidateList(context.Background(), validator.ListRequest{
				Path:        path.Root("container_reg_ids"),
				ConfigValue: tt.value,
			}, resp)

			if resp.Diagnostics.HasError() != tt.expectError {
				t.Fatalf("expected error=%v, got diagnostics: %v", tt.expectError, resp.Diagnostics)
			}
		})
	}
}
