package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContainerRegistryCredentialIDsResourceSchemas(t *testing.T) {
	resources := []struct {
		name     string
		resource resource.Resource
		parents  []string
	}{
		{"azure_batch_ce", &AzureBatchCEResource{}, []string{"config", "forge"}},
		{"compute_env", &ComputeEnvResource{}, []string{"compute_env", "config", "azure_batch", "forge"}},
	}
	list := func(values ...attr.Value) types.List {
		return types.ListValueMust(types.StringType, values)
	}
	cases := []struct {
		name    string
		value   types.List
		invalid bool
	}{
		{"null", types.ListNull(types.StringType), false},
		{"unknown", types.ListUnknown(types.StringType), false},
		{"empty", list(), false},
		{"credential ID", list(types.StringValue("2dzgMSW1UqQHFhZqGQZ99X")), false},
		{"unknown ID", list(types.StringUnknown()), false},
		{"multiple IDs", list(types.StringValue("one"), types.StringValue("two")), true},
		{"Azure resource ID", list(types.StringValue("/subscriptions/example/registries/acr")), true},
		{"empty ID", list(types.StringValue("")), true},
		{"null ID", list(types.StringNull()), true},
	}
	for _, r := range resources {
		t.Run(r.name, func(t *testing.T) {
			var response resource.SchemaResponse
			r.resource.Schema(context.Background(), resource.SchemaRequest{}, &response)
			require.False(t, response.Diagnostics.HasError())
			attributes := response.Schema.Attributes
			for _, parent := range r.parents {
				nested, ok := attributes[parent].(schema.SingleNestedAttribute)
				require.True(t, ok, "expected nested attribute %s", parent)
				attributes = nested.Attributes
			}
			ids, ok := attributes["container_reg_ids"].(schema.ListAttribute)
			require.True(t, ok)
			require.NotEmpty(t, ids.Validators)
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					var validation validator.ListResponse
					for _, v := range ids.Validators {
						v.ValidateList(context.Background(), validator.ListRequest{
							Path: path.Root("container_reg_ids"), ConfigValue: tc.value,
						}, &validation)
					}
					require.Equal(t, tc.invalid, validation.Diagnostics.HasError(), "%v", validation.Diagnostics)
				})
			}
		})
	}
}
