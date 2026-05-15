package stringvalidators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRunScriptSizeValidator(t *testing.T) {
	tests := []struct {
		name          string
		value         types.String
		expectWarning bool
	}{
		{
			name:          "null value skipped",
			value:         types.StringNull(),
			expectWarning: false,
		},
		{
			name:          "unknown value skipped",
			value:         types.StringUnknown(),
			expectWarning: false,
		},
		{
			name:          "empty string is well under limit",
			value:         types.StringValue(""),
			expectWarning: false,
		},
		{
			name:          "short script is under limit",
			value:         types.StringValue("echo hello"),
			expectWarning: false,
		},
		{
			name:          "exactly at limit is allowed",
			value:         types.StringValue(strings.Repeat("x", runScriptSizeSoftLimit)),
			expectWarning: false,
		},
		{
			name:          "one byte over limit warns",
			value:         types.StringValue(strings.Repeat("x", runScriptSizeSoftLimit+1)),
			expectWarning: true,
		},
		{
			name:          "well over limit warns",
			value:         types.StringValue(strings.Repeat("x", runScriptSizeSoftLimit*4)),
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &validator.StringResponse{Diagnostics: diag.Diagnostics{}}
			RunScriptSizeValidator().ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("pre_run_script"),
				ConfigValue: tt.value,
			}, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("validator emitted an Error; expected Warning only. diags: %v", resp.Diagnostics)
			}

			warnings := resp.Diagnostics.Warnings()
			gotWarning := len(warnings) > 0
			if gotWarning != tt.expectWarning {
				t.Errorf("expected warning=%v, got warning=%v (diags: %v)", tt.expectWarning, gotWarning, resp.Diagnostics)
			}
		})
	}
}
