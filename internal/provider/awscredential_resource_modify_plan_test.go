package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/stretchr/testify/assert"
)

const testRoleArn = "arn:aws:iam::123456789012:role/SeqeraRole"

func awsCredentialSchema(t *testing.T) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	(&AWSCredentialResource{}).Schema(context.Background(), resource.SchemaRequest{}, resp)
	return resp.Schema
}

// stubInstanceCredentials replaces the /service-info lookup for one test and counts the calls.
func stubInstanceCredentials(t *testing.T, allowed, known bool) *int {
	t.Helper()
	calls := 0
	previous := instanceCredentialsAllowed
	instanceCredentialsAllowed = func(context.Context, *sdk.Seqera) (bool, bool) {
		calls++
		return allowed, known
	}
	t.Cleanup(func() { instanceCredentialsAllowed = previous })
	return &calls
}

// modifyAWSCredentialPlan runs ModifyPlan for a create (state == nil) or an update.
func modifyAWSCredentialPlan(t *testing.T, client *sdk.Seqera, config, state map[string]tftypes.Value) diag.Diagnostics {
	t.Helper()
	s := awsCredentialSchema(t)
	base := map[string]tftypes.Value{"name": str("aws"), "workspace_id": tftypes.NewValue(tftypes.Number, 1)}
	for k, v := range config {
		base[k] = v
	}
	var stateVals map[string]tftypes.Value
	if state != nil {
		stateVals = map[string]tftypes.Value{"name": str("aws"), "workspace_id": tftypes.NewValue(tftypes.Number, 1)}
		for k, v := range state {
			stateVals[k] = v
		}
	}
	f := newPlanFixture(t, s, stateVals, base, base)
	if state == nil {
		f.state = tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
	}
	req := resource.ModifyPlanRequest{Config: f.config, Plan: f.plan, State: f.state}
	resp := &resource.ModifyPlanResponse{Plan: f.plan}
	(&AWSCredentialResource{client: client}).ModifyPlan(context.Background(), req, resp)
	return resp.Diagnostics
}

func TestAWSCredentialModifyPlan_UnsetModeWithRoleArnOnlyFailsWithoutInstanceCredentials(t *testing.T) {
	stubInstanceCredentials(t, false, true)

	diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, map[string]tftypes.Value{"assume_role_arn": str(testRoleArn)}, nil)

	assert.True(t, diags.HasError(), "the Platform treats an unset mode as keys and rejects it without keys")
	assert.Contains(t, diags.Errors()[0].Detail(), `unset or "keys"`)
}

func TestAWSCredentialModifyPlan_UnsetModeWithRoleArnOnlyAllowedWithInstanceCredentials(t *testing.T) {
	stubInstanceCredentials(t, true, true)

	diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, map[string]tftypes.Value{"assume_role_arn": str(testRoleArn)}, nil)

	assert.False(t, diags.HasError(), "with instance credentials the role is assumed with the instance identity, got: %s", diags.Errors())
}

func TestAWSCredentialModifyPlan_RoleModeWithKeys(t *testing.T) {
	config := map[string]tftypes.Value{
		"mode":            str("role"),
		"assume_role_arn": str(testRoleArn),
		"access_key":      str("AKIAIOSFODNN7EXAMPLE"),
		"secret_key":      str("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
	}

	t.Run("rejected without instance credentials", func(t *testing.T) {
		stubInstanceCredentials(t, false, true)
		diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, config, nil)
		assert.True(t, diags.HasError())
		assert.Equal(t, "Conflicting Attributes", diags.Errors()[0].Summary())
	})

	t.Run("accepted with instance credentials, which the Platform does not check", func(t *testing.T) {
		stubInstanceCredentials(t, true, true)
		diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, config, nil)
		assert.False(t, diags.HasError(), "got: %s", diags.Errors())
	})

	t.Run("accepted when the setting cannot be read", func(t *testing.T) {
		stubInstanceCredentials(t, false, false)
		diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, config, nil)
		assert.False(t, diags.HasError(), "got: %s", diags.Errors())
	})
}

func TestAWSCredentialModifyPlan_WorkloadIdentityRejectsUseExternalID(t *testing.T) {
	stubInstanceCredentials(t, false, true)

	diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, map[string]tftypes.Value{
		"mode":            str("workloadIdentity"),
		"assume_role_arn": str(testRoleArn),
		"use_external_id": tftypes.NewValue(tftypes.Bool, true),
	}, nil)

	assert.True(t, diags.HasError())
	errWithPath, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
	assert.True(t, ok)
	assert.Equal(t, path.Root("use_external_id"), errWithPath.Path())
}

func TestAWSCredentialModifyPlan_UnsetModeOnUpdateUsesStoredMode(t *testing.T) {
	stubInstanceCredentials(t, false, true)

	// mode removed from the config of an existing role credential: the stored mode is kept and sent
	diags := modifyAWSCredentialPlan(t, &sdk.Seqera{},
		map[string]tftypes.Value{"assume_role_arn": str(testRoleArn)},
		map[string]tftypes.Value{"mode": str("role"), "assume_role_arn": str(testRoleArn)})

	assert.False(t, diags.HasError(), "got: %s", diags.Errors())
}

func TestAWSCredentialModifyPlan_UnchangedLegacyCredentialsPlanCleanly(t *testing.T) {
	calls := stubInstanceCredentials(t, false, true)

	// Created before modes existed: a role ARN, no mode, no keys. Planning with no change must not fail.
	legacy := map[string]tftypes.Value{"assume_role_arn": str(testRoleArn)}
	diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, legacy, legacy)

	assert.False(t, diags.HasError(), "a plan that changes nothing must not be blocked, got: %s", diags.Errors())
	assert.Equal(t, 0, *calls, "no Platform lookup when nothing changes")
}

func TestAWSCredentialModifyPlan_ChangedLegacyCredentialsAreChecked(t *testing.T) {
	stubInstanceCredentials(t, false, true)

	// The same legacy credentials, now changed: the Platform treats the missing mode as keys and rejects it
	diags := modifyAWSCredentialPlan(t, &sdk.Seqera{},
		map[string]tftypes.Value{"assume_role_arn": str("arn:aws:iam::123456789012:role/OtherRole")},
		map[string]tftypes.Value{"assume_role_arn": str(testRoleArn)})

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), `unset or "keys"`)
}

func TestAWSCredentialModifyPlan_SkipsUnknownValues(t *testing.T) {
	calls := stubInstanceCredentials(t, false, true)

	diags := modifyAWSCredentialPlan(t, &sdk.Seqera{}, map[string]tftypes.Value{
		"mode":            str("role"),
		"assume_role_arn": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	}, nil)

	assert.False(t, diags.HasError(), "got: %s", diags.Errors())
	assert.Equal(t, 0, *calls, "no Platform lookup for values only known at apply time")
}

func TestAWSCredentialModifyPlan_SkipsUnconfiguredProvider(t *testing.T) {
	calls := stubInstanceCredentials(t, false, true)

	diags := modifyAWSCredentialPlan(t, nil, map[string]tftypes.Value{"assume_role_arn": str(testRoleArn)}, nil)

	assert.False(t, diags.HasError(), "got: %s", diags.Errors())
	assert.Equal(t, 0, *calls)
}

func TestAWSCredentialModifyPlan_SkipsDestroy(t *testing.T) {
	calls := stubInstanceCredentials(t, false, true)
	s := awsCredentialSchema(t)
	nullObject := tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)

	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config{Schema: s, Raw: nullObject},
		Plan:   tfsdk.Plan{Schema: s, Raw: nullObject},
		State:  tfsdk.State{Schema: s, Raw: objectValue(t, s, map[string]tftypes.Value{"mode": str("role")})},
	}
	resp := &resource.ModifyPlanResponse{Plan: req.Plan}
	(&AWSCredentialResource{client: &sdk.Seqera{}}).ModifyPlan(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "got: %s", resp.Diagnostics.Errors())
	assert.Equal(t, 0, *calls)
}

func TestAWSCredentialModeDiagnostics(t *testing.T) {
	cases := []struct {
		name      string
		fields    awsCredentialFields
		wantPath  string // empty when the fields are valid
		wantInMsg string
	}{
		{"keys mode with keys", awsCredentialFields{Mode: "keys", AccessKey: true, SecretKey: true}, "", ""},
		{"unset mode with keys and a role ARN", awsCredentialFields{AccessKey: true, SecretKey: true, AssumeRoleArn: true}, "", ""},
		{"keys mode with only a role ARN", awsCredentialFields{Mode: "keys", AssumeRoleArn: true}, "access_key", `unset or "keys"`},
		{"role mode with a role ARN", awsCredentialFields{Mode: "role", AssumeRoleArn: true, UseExternalID: true}, "", ""},
		{"role mode without a role ARN", awsCredentialFields{Mode: "role"}, "assume_role_arn", `'mode' is "role"`},
		{"workload identity with a role ARN", awsCredentialFields{Mode: "workloadIdentity", AssumeRoleArn: true}, "", ""},
		{"workload identity with only a secret key", awsCredentialFields{Mode: "workloadIdentity", SecretKey: true, AssumeRoleArn: true}, "secret_key", `"workloadIdentity"`},
		{"workload identity with use_external_id", awsCredentialFields{Mode: "workloadIdentity", AssumeRoleArn: true, UseExternalID: true}, "use_external_id", "External ID"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := awsCredentialModeDiagnostics(tc.fields)
			if tc.wantPath == "" {
				assert.False(t, diags.HasError(), "got: %s", diags.Errors())
				return
			}
			assert.True(t, diags.HasError())
			errWithPath, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
			assert.True(t, ok)
			assert.Equal(t, path.Root(tc.wantPath), errWithPath.Path())
			assert.Contains(t, diags.Errors()[0].Detail(), tc.wantInMsg)
		})
	}
}
