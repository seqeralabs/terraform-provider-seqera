package stringvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWorkDirFormatValidator(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		// Valid cases
		{
			name:        "valid s3 path with subpath",
			value:       types.StringValue("s3://my-bucket/work"),
			expectError: false,
		},
		{
			name:        "valid s3 bucket only",
			value:       types.StringValue("s3://my-bucket"),
			expectError: false,
		},
		{
			name:        "valid gs path",
			value:       types.StringValue("gs://my-bucket/work"),
			expectError: false,
		},
		{
			name:        "valid az path",
			value:       types.StringValue("az://my-container/work"),
			expectError: false,
		},
		{
			name:        "valid local absolute path",
			value:       types.StringValue("/scratch/work"),
			expectError: false,
		},
		{
			name:        "valid root path",
			value:       types.StringValue("/"),
			expectError: false,
		},
		{
			name:        "null value skipped",
			value:       types.StringNull(),
			expectError: false,
		},
		{
			name:        "unknown value skipped",
			value:       types.StringUnknown(),
			expectError: false,
		},
		{
			name:        "empty string skipped",
			value:       types.StringValue(""),
			expectError: false,
		},
		// Invalid prefix cases
		{
			name:        "invalid relative path",
			value:       types.StringValue("work/dir"),
			expectError: true,
		},
		{
			name:        "invalid http URL",
			value:       types.StringValue("http://bucket/work"),
			expectError: true,
		},
		{
			name:        "invalid ftp URL",
			value:       types.StringValue("ftp://bucket/work"),
			expectError: true,
		},
		{
			name:        "invalid hdfs path",
			value:       types.StringValue("hdfs://bucket/work"),
			expectError: true,
		},
		// Trailing slash cases
		{
			name:        "trailing slash on s3 path",
			value:       types.StringValue("s3://my-bucket/work/"),
			expectError: true,
		},
		{
			name:        "trailing slash on gs path",
			value:       types.StringValue("gs://my-bucket/work/"),
			expectError: true,
		},
		{
			name:        "trailing slash on az path",
			value:       types.StringValue("az://my-container/work/"),
			expectError: true,
		},
		{
			name:        "trailing slash on local path",
			value:       types.StringValue("/scratch/work/"),
			expectError: true,
		},
		// Missing bucket/container name
		{
			name:        "s3 prefix only",
			value:       types.StringValue("s3://"),
			expectError: true,
		},
		{
			name:        "gs prefix only",
			value:       types.StringValue("gs://"),
			expectError: true,
		},
		{
			name:        "az prefix only",
			value:       types.StringValue("az://"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WorkDirFormatValidator()
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}
			v.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("work_dir"),
				ConfigValue: tt.value,
			}, resp)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error for value %q, but got none", tt.value)
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("expected no error for value %q, but got: %s", tt.value, resp.Diagnostics.Errors()[0].Detail())
			}
		})
	}
}
