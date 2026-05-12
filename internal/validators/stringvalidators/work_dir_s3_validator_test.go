package stringvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWorkDirS3Validator(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		// Valid
		{name: "s3 path with subpath", value: types.StringValue("s3://my-bucket/work"), expectError: false},
		{name: "s3 bucket only", value: types.StringValue("s3://my-bucket"), expectError: false},
		{name: "null skipped", value: types.StringNull(), expectError: false},
		{name: "unknown skipped", value: types.StringUnknown(), expectError: false},
		{name: "empty string skipped", value: types.StringValue(""), expectError: false},

		// Invalid prefix
		{name: "gs path rejected", value: types.StringValue("gs://my-bucket/work"), expectError: true},
		{name: "az path rejected", value: types.StringValue("az://my-container/work"), expectError: true},
		{name: "local absolute rejected", value: types.StringValue("/scratch/work"), expectError: true},
		{name: "relative path rejected", value: types.StringValue("work/dir"), expectError: true},
		{name: "http URL rejected", value: types.StringValue("http://bucket/work"), expectError: true},
		{name: "double-slash typo rejected", value: types.StringValue("//s3"), expectError: true},
		{name: "S3 uppercase rejected", value: types.StringValue("S3://my-bucket"), expectError: true},

		// Missing bucket
		{name: "s3 prefix only", value: types.StringValue("s3://"), expectError: true},

		// Trailing slash
		{name: "trailing slash on s3 path", value: types.StringValue("s3://my-bucket/work/"), expectError: true},
		{name: "trailing slash on bucket", value: types.StringValue("s3://my-bucket/"), expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := WorkDirS3Validator()
			resp := &validator.StringResponse{Diagnostics: diag.Diagnostics{}}
			v.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("work_dir"),
				ConfigValue: tt.value,
			}, resp)

			if tt.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("expected error for value %q, but got none", tt.value)
			}
			if !tt.expectError && resp.Diagnostics.HasError() {
				t.Errorf("expected no error for value %q, but got: %v", tt.value, resp.Diagnostics)
			}
		})
	}
}
