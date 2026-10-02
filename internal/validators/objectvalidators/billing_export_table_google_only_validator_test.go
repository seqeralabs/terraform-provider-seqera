package objectvalidators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// makeBillingExportTableRequest builds a validator.ObjectRequest for a `config`
// object carrying a nested intelligent_compute_config.billing_export_table, the
// shape the validator sees when attached to AwsCloudConfig / AzCloudConfig /
// LocalComputeConfig. A nil table means the attribute is null; unknownTable
// overrides it with an unknown value.
func makeBillingExportTableRequest(t *testing.T, schedConfigNull bool, table *string, unknownTable bool) validator.ObjectRequest {
	t.Helper()

	schedType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"billing_export_table": tftypes.String,
		},
	}
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"intelligent_compute_config": schedType,
		},
	}

	var tableRaw tftypes.Value
	switch {
	case unknownTable:
		tableRaw = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	case table == nil:
		tableRaw = tftypes.NewValue(tftypes.String, nil)
	default:
		tableRaw = tftypes.NewValue(tftypes.String, *table)
	}

	schedRaw := tftypes.NewValue(schedType, map[string]tftypes.Value{"billing_export_table": tableRaw})
	if schedConfigNull {
		schedRaw = tftypes.NewValue(schedType, nil)
	}
	configRaw := tftypes.NewValue(configType, map[string]tftypes.Value{"intelligent_compute_config": schedRaw})
	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"config": configType}}
	rootRaw := tftypes.NewValue(rootType, map[string]tftypes.Value{"config": configRaw})

	schedAttr := resourceschema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]resourceschema.Attribute{
			"billing_export_table": resourceschema.StringAttribute{Optional: true},
		},
	}
	schema := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"config": resourceschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]resourceschema.Attribute{
					"intelligent_compute_config": schedAttr,
				},
			},
		},
	}

	schedAttrTypes := map[string]attr.Type{"billing_export_table": types.StringType}

	schedValue := types.ObjectNull(schedAttrTypes)
	if !schedConfigNull {
		var tableValue attr.Value
		switch {
		case unknownTable:
			tableValue = types.StringUnknown()
		case table == nil:
			tableValue = types.StringNull()
		default:
			tableValue = types.StringValue(*table)
		}
		v, d := types.ObjectValue(schedAttrTypes, map[string]attr.Value{"billing_export_table": tableValue})
		if d.HasError() {
			t.Fatalf("building intelligent_compute_config: %v", d)
		}
		schedValue = v
	}

	configValue, diags := types.ObjectValue(
		map[string]attr.Type{"intelligent_compute_config": schedValue.Type(context.Background())},
		map[string]attr.Value{"intelligent_compute_config": schedValue},
	)
	if diags.HasError() {
		t.Fatalf("building config object: %v", diags)
	}

	return validator.ObjectRequest{
		Path:        path.Root("config"),
		ConfigValue: configValue,
		Config:      tfsdk.Config{Schema: schema, Raw: rootRaw},
	}
}

func TestBillingExportTableGoogleOnlyValidator(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name            string
		schedConfigNull bool
		table           *string
		unknownTable    bool
		expectError     bool
	}{
		{
			name:            "no intelligent_compute_config passes",
			schedConfigNull: true,
		},
		{
			name: "unset billing_export_table passes",
		},
		{
			name:  "empty billing_export_table passes",
			table: str(""),
		},
		{
			name:         "unknown billing_export_table passes",
			unknownTable: true,
		},
		{
			name:        "configured billing_export_table is rejected",
			table:       str("my-project.billing.gcp_billing_export_v1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := makeBillingExportTableRequest(t, tt.schedConfigNull, tt.table, tt.unknownTable)
			resp := &validator.ObjectResponse{}

			BillingExportTableGoogleOnlyValidator().ValidateObject(context.Background(), req, resp)

			if tt.expectError != resp.Diagnostics.HasError() {
				t.Fatalf("expectError=%v, got diagnostics: %v", tt.expectError, resp.Diagnostics)
			}
			if !tt.expectError {
				return
			}
			if got := resp.Diagnostics.Errors()[0].Summary(); got != "Unsupported billing_export_table" {
				t.Errorf("unexpected summary: %q", got)
			}
			if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, "Google Cloud") {
				t.Errorf("detail should point at Google Cloud, got: %q", detail)
			}
		})
	}
}
