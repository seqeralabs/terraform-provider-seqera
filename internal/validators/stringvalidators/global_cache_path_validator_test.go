package stringvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestGlobalCachePathValidator(t *testing.T) {
	t.Parallel()

	sch := schema.Schema{Attributes: map[string]schema.Attribute{
		"global_cache_path":    schema.StringAttribute{Optional: true},
		"global_cache_enabled": schema.BoolAttribute{Optional: true},
	}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"global_cache_path":    tftypes.String,
		"global_cache_enabled": tftypes.Bool,
	}}

	enabledTrue := tftypes.NewValue(tftypes.Bool, true)

	tests := map[string]struct {
		validator validator.String
		path      types.String
		enabled   tftypes.Value
		wantError bool
	}{
		"path null, disabled":       {GlobalCachePathS3Validator(), types.StringNull(), tftypes.NewValue(tftypes.Bool, false), false},
		"path unknown":              {GlobalCachePathS3Validator(), types.StringUnknown(), tftypes.NewValue(tftypes.Bool, nil), false},
		"valid s3, enabled":         {GlobalCachePathS3Validator(), types.StringValue("s3://bucket/cache"), enabledTrue, false},
		"valid bucket only":         {GlobalCachePathS3Validator(), types.StringValue("s3://bucket"), enabledTrue, false},
		"valid s3, enabled unknown": {GlobalCachePathS3Validator(), types.StringValue("s3://bucket/cache"), tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), false},
		"valid gs":                  {GlobalCachePathGsValidator(), types.StringValue("gs://bucket/cache"), enabledTrue, false},
		"valid az":                  {GlobalCachePathAzValidator(), types.StringValue("az://container/cache"), enabledTrue, false},
		"path set, enabled null":    {GlobalCachePathS3Validator(), types.StringValue("s3://bucket/cache"), tftypes.NewValue(tftypes.Bool, nil), true},
		"path set, enabled false":   {GlobalCachePathS3Validator(), types.StringValue("s3://bucket/cache"), tftypes.NewValue(tftypes.Bool, false), true},
		"no scheme":                 {GlobalCachePathS3Validator(), types.StringValue("bucket/cache"), enabledTrue, true},
		"wrong scheme":              {GlobalCachePathS3Validator(), types.StringValue("gs://bucket/cache"), enabledTrue, true},
		"trailing slash":            {GlobalCachePathS3Validator(), types.StringValue("s3://bucket/cache/"), enabledTrue, true},
		"repeated slashes":          {GlobalCachePathGsValidator(), types.StringValue("gs://bucket//cache"), enabledTrue, true},
		"scheme only":               {GlobalCachePathAzValidator(), types.StringValue("az://"), enabledTrue, true},
		"scheme with slashes only":  {GlobalCachePathAzValidator(), types.StringValue("az:///"), enabledTrue, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var pathVal tftypes.Value
			switch {
			case tc.path.IsNull():
				pathVal = tftypes.NewValue(tftypes.String, nil)
			case tc.path.IsUnknown():
				pathVal = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			default:
				pathVal = tftypes.NewValue(tftypes.String, tc.path.ValueString())
			}

			req := validator.StringRequest{
				Path:        path.Root("global_cache_path"),
				ConfigValue: tc.path,
				Config: tfsdk.Config{
					Schema: sch,
					Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
						"global_cache_path":    pathVal,
						"global_cache_enabled": tc.enabled,
					}),
				},
			}
			resp := &validator.StringResponse{}
			tc.validator.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("wantError=%v, diagnostics=%v", tc.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestNormaliseGlobalCachePath(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"s3://bucket/cache":    "s3://bucket/cache",
		"s3://bucket/cache/":   "s3://bucket/cache",
		"s3://bucket//a///b//": "s3://bucket/a/b",
		"gs://bucket":          "gs://bucket",
		"no-scheme/path/":      "no-scheme/path/",
	}
	for in, want := range tests {
		if got := normaliseGlobalCachePath(in); got != want {
			t.Errorf("normaliseGlobalCachePath(%q) = %q, want %q", in, got, want)
		}
	}
}
