package stringvalidators

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// awsCredObjectType returns the tftypes for the AWS credential config block.
func awsCredObjectType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"access_key":      tftypes.String,
			"secret_key":      tftypes.String,
			"assume_role_arn": tftypes.String,
			"mode":            tftypes.String,
			"use_external_id": tftypes.Bool,
		},
	}
}

// awsCredSchema returns the sibling attributes the AWS credential validator reads.
func awsCredSchema() resourceschema.Schema {
	return resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"access_key":      resourceschema.StringAttribute{Optional: true},
			"secret_key":      resourceschema.StringAttribute{Optional: true},
			"assume_role_arn": resourceschema.StringAttribute{Optional: true},
			"mode":            resourceschema.StringAttribute{Optional: true},
			"use_external_id": resourceschema.BoolAttribute{Optional: true},
		},
	}
}

// makeAWSRequest constructs a validator.StringRequest for the AWS credential validator.
// fieldName is the field being validated (access_key, secret_key, or assume_role_arn).
// values maps field names to string values; nil means null. use_external_id takes "true" or "false".
func makeAWSRequest(fieldName string, values map[string]*string) validator.StringRequest {
	ct := awsCredObjectType()

	tfvals := make(map[string]tftypes.Value)
	for _, name := range []string{"access_key", "secret_key", "assume_role_arn", "mode"} {
		if v, ok := values[name]; ok && v != nil {
			tfvals[name] = tftypes.NewValue(tftypes.String, *v)
		} else {
			tfvals[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	if v, ok := values["use_external_id"]; ok && v != nil {
		tfvals["use_external_id"] = tftypes.NewValue(tftypes.Bool, *v == "true")
	} else {
		tfvals["use_external_id"] = tftypes.NewValue(tftypes.Bool, nil)
	}

	rawVal := tftypes.NewValue(ct, tfvals)

	s := awsCredSchema()

	config := tfsdk.Config{
		Schema: s,
		Raw:    rawVal,
	}

	var configValue types.String
	if v, ok := values[fieldName]; ok && v != nil {
		configValue = types.StringValue(*v)
	} else {
		configValue = types.StringNull()
	}

	return validator.StringRequest{
		Path:        path.Root(fieldName),
		ConfigValue: configValue,
		Config:      config,
	}
}

func strPtr(s string) *string {
	return &s
}

func runAWSValidator(req validator.StringRequest) diag.Diagnostics {
	return runStringValidator(AWSCredentialKeysValidator(), req)
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------

func TestAWSCredentialKeys_BothKeysProvided(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"access_key":      strPtr("AKIAIOSFODNN7EXAMPLE"),
		"secret_key":      strPtr("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		"assume_role_arn": nil,
	}
	diags := runAWSValidator(makeAWSRequest("access_key", values))
	assert.False(t, diags.HasError(), "both access_key and secret_key should be valid, got: %s", diags.Errors())
}

func TestAWSCredentialKeys_AssumeRoleOnly(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"access_key":      nil,
		"secret_key":      nil,
		"assume_role_arn": strPtr("arn:aws:iam::123456789012:role/MyRole"),
	}
	diags := runAWSValidator(makeAWSRequest("assume_role_arn", values))
	assert.False(t, diags.HasError(), "assume_role_arn alone should be valid, got: %s", diags.Errors())
}

func TestAWSCredentialKeys_AllThreeProvided(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"access_key":      strPtr("AKIAIOSFODNN7EXAMPLE"),
		"secret_key":      strPtr("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		"assume_role_arn": strPtr("arn:aws:iam::123456789012:role/MyRole"),
	}
	diags := runAWSValidator(makeAWSRequest("access_key", values))
	assert.False(t, diags.HasError(), "all three fields should be valid, got: %s", diags.Errors())
}

func TestAWSCredentialKeys_AccessKeyWithoutSecretKey(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"access_key":      strPtr("AKIAIOSFODNN7EXAMPLE"),
		"secret_key":      nil,
		"assume_role_arn": nil,
	}
	diags := runAWSValidator(makeAWSRequest("access_key", values))
	assert.True(t, diags.HasError(), "access_key without secret_key should fail")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Attribute")
}

func TestAWSCredentialKeys_SecretKeyWithoutAccessKey(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"access_key":      nil,
		"secret_key":      strPtr("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		"assume_role_arn": nil,
	}
	diags := runAWSValidator(makeAWSRequest("secret_key", values))
	assert.True(t, diags.HasError(), "secret_key without access_key should fail")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Attribute")
}

func TestAWSCredentialKeys_NothingProvided(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"access_key":      nil,
		"secret_key":      nil,
		"assume_role_arn": nil,
	}
	diags := runAWSValidator(makeAWSRequest("access_key", values))
	assert.True(t, diags.HasError(), "no credentials at all should fail")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Configuration")
}

func TestAWSCredentialKeys_UnknownValuesSkipsValidation(t *testing.T) {
	t.Parallel()
	ct := awsCredObjectType()
	rawVal := tftypes.NewValue(ct, map[string]tftypes.Value{
		"access_key":      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"secret_key":      tftypes.NewValue(tftypes.String, nil),
		"assume_role_arn": tftypes.NewValue(tftypes.String, nil),
		"mode":            tftypes.NewValue(tftypes.String, nil),
		"use_external_id": tftypes.NewValue(tftypes.Bool, nil),
	})

	s := awsCredSchema()

	req := validator.StringRequest{
		Path:        path.Root("access_key"),
		ConfigValue: types.StringUnknown(),
		Config:      tfsdk.Config{Schema: s, Raw: rawVal},
	}

	diags := runAWSValidator(req)
	assert.False(t, diags.HasError(), "unknown values should skip validation, got: %s", diags.Errors())
}

func TestAWSCredentialKeys_EmptyStringsCountAsUnset(t *testing.T) {
	t.Parallel()
	values := map[string]*string{
		"access_key":      strPtr(""),
		"secret_key":      strPtr(""),
		"assume_role_arn": strPtr(""),
	}
	diags := runAWSValidator(makeAWSRequest("access_key", values))
	assert.True(t, diags.HasError(), "empty strings should be treated as unset")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Configuration")
}

func TestAWSCredentialKeys_AssumeRoleWithKeys(t *testing.T) {
	t.Parallel()
	// Using assume_role_arn together with access_key+secret_key is valid
	// (cross-account role assumption with explicit credentials).
	values := map[string]*string{
		"access_key":      strPtr("AKIAIOSFODNN7EXAMPLE"),
		"secret_key":      strPtr("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		"assume_role_arn": strPtr("arn:aws:iam::123456789012:role/MyRole"),
	}

	// Validate from each field's perspective — all should pass.
	for _, field := range []string{"access_key", "secret_key", "assume_role_arn"} {
		t.Run("validate_"+field, func(t *testing.T) {
			t.Parallel()
			diags := runAWSValidator(makeAWSRequest(field, values))
			assert.False(t, diags.HasError(), "all three fields valid when validated from %s, got: %s", field, diags.Errors())
		})
	}
}

func TestAWSCredentialKeys_AssumeRoleWithAccessKeyOnly(t *testing.T) {
	t.Parallel()
	// assume_role_arn + access_key but no secret_key should fail.
	values := map[string]*string{
		"access_key":      strPtr("AKIAIOSFODNN7EXAMPLE"),
		"secret_key":      nil,
		"assume_role_arn": strPtr("arn:aws:iam::123456789012:role/MyRole"),
	}
	diags := runAWSValidator(makeAWSRequest("access_key", values))
	assert.True(t, diags.HasError(), "access_key without secret_key should fail even with assume_role_arn")
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing Required Attribute")
}

func TestAWSCredentialKeys_NothingProvidedReportedOnce(t *testing.T) {
	t.Parallel()
	// The validator runs on three attributes; the error must attach to one fixed path so it is reported once.
	for _, field := range []string{"access_key", "secret_key", "assume_role_arn"} {
		diags := runAWSValidator(makeAWSRequest(field, map[string]*string{}))
		assert.Len(t, diags.Errors(), 1, "validating %s", field)
		errWithPath, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
		assert.True(t, ok, "validating %s: expected an attribute error", field)
		assert.Equal(t, path.Root("assume_role_arn"), errWithPath.Path(), "validating %s", field)
	}
}

func TestAWSCredentialKeys_IgnoresMode(t *testing.T) {
	t.Parallel()
	// Mode rules depend on the Platform's instance-credentials setting and live in the resource's ModifyPlan.
	values := map[string]*string{
		"access_key":      strPtr("AKIAIOSFODNN7EXAMPLE"),
		"secret_key":      strPtr("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		"assume_role_arn": strPtr("arn:aws:iam::123456789012:role/MyRole"),
		"mode":            strPtr("role"),
	}
	diags := runAWSValidator(makeAWSRequest("access_key", values))
	assert.False(t, diags.HasError(), "the validator must not apply mode rules, got: %s", diags.Errors())
}
