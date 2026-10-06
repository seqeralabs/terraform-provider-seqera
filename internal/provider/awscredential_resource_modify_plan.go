package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
)

var _ resource.ResourceWithModifyPlan = &AWSCredentialResource{}

// ModifyPlan checks the fields each AWS credential mode allows, mirroring the Platform's
// AwsSecurityKeys.validate(), so a mismatch fails at plan time instead of at apply.
//
// The Platform skips those rules when it allows instance credentials (TOWER_ALLOW_INSTANCE_CREDENTIALS,
// reported as allowInstanceCredentials by /service-info): there, for example, a role ARN without keys is
// assumed with the instance's own identity whatever the mode. The rules are therefore applied only when the
// Platform reports that it does not allow instance credentials.
func (r *AWSCredentialResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to check on destroy, or before the provider is configured
	if req.Plan.Raw.IsNull() || r.client == nil {
		return
	}

	// Nothing to check when existing credentials do not change. Credentials created before modes existed can
	// hold a role ARN with no mode and no keys; checking them on every plan would block plans that change
	// nothing, while the rules still apply as soon as they are created or updated.
	if !req.State.Raw.IsNull() && req.Plan.Raw.Equal(req.State.Raw) {
		return
	}

	var mode, accessKey, secretKey, assumeRoleArn types.String
	var useExternalID types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("mode"), &mode)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("access_key"), &accessKey)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("secret_key"), &secretKey)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("assume_role_arn"), &assumeRoleArn)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("use_external_id"), &useExternalID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Values only known at apply time cannot be checked yet
	if mode.IsUnknown() || accessKey.IsUnknown() || secretKey.IsUnknown() || assumeRoleArn.IsUnknown() || useExternalID.IsUnknown() {
		return
	}

	// An unset mode keeps the existing credentials' mode, which is the one sent to the Platform. For new
	// credentials it stays unset, which the Platform treats as keys.
	if mode.IsNull() && !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("mode"), &mode)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	allowed, known := instanceCredentialsAllowed(ctx, r.client)
	if !known || allowed {
		return
	}

	resp.Diagnostics.Append(awsCredentialModeDiagnostics(awsCredentialFields{
		Mode:          mode.ValueString(),
		AccessKey:     awsFieldProvided(accessKey),
		SecretKey:     awsFieldProvided(secretKey),
		AssumeRoleArn: awsFieldProvided(assumeRoleArn),
		UseExternalID: !useExternalID.IsNull() && useExternalID.ValueBool(),
	})...)
}

// awsCredentialFields are the AWS credential settings the mode rules look at.
type awsCredentialFields struct {
	Mode          string
	AccessKey     bool
	SecretKey     bool
	AssumeRoleArn bool
	UseExternalID bool
}

// awsCredentialModeDiagnostics reports the fields the given mode does not allow or requires, as the Platform's
// AwsSecurityKeys.validate() does. Pairing access_key with secret_key, and requiring some way to authenticate,
// hold on every installation and are checked by AWSCredentialKeysValidator.
func awsCredentialModeDiagnostics(f awsCredentialFields) diag.Diagnostics {
	var diags diag.Diagnostics

	switch f.Mode {
	case "role", "workloadIdentity":
		if f.AccessKey || f.SecretKey {
			keyPath := path.Root("access_key")
			if !f.AccessKey {
				keyPath = path.Root("secret_key")
			}
			diags.AddAttributeError(
				keyPath,
				"Conflicting Attributes",
				fmt.Sprintf("The 'access_key' and 'secret_key' attributes must not be set when 'mode' is %q. This mode authenticates by assuming 'assume_role_arn' without static keys.", f.Mode),
			)
		}
		if !f.AssumeRoleArn {
			diags.AddAttributeError(
				path.Root("assume_role_arn"),
				"Missing Required Attribute",
				fmt.Sprintf("The 'assume_role_arn' attribute is required when 'mode' is %q.", f.Mode),
			)
		}
		if f.Mode == "workloadIdentity" && f.UseExternalID {
			diags.AddAttributeError(
				path.Root("use_external_id"),
				"Conflicting Attributes",
				"The 'use_external_id' attribute must not be true when 'mode' is \"workloadIdentity\". Workload identity federation does not use an External ID.",
			)
		}
	default:
		// "keys", or unset: the Platform treats a missing mode as keys
		if !f.AccessKey && !f.SecretKey {
			diags.AddAttributeError(
				path.Root("access_key"),
				"Missing Required Attribute",
				"The 'access_key' and 'secret_key' attributes are required when 'mode' is unset or \"keys\". To assume 'assume_role_arn' without static keys, set 'mode' to \"role\" or \"workloadIdentity\".",
			)
		}
	}

	return diags
}

func awsFieldProvided(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown() && v.ValueString() != ""
}

// instanceCredentialsLookup caches one /service-info lookup per configured client, so a plan with many AWS
// credentials asks the Platform once.
type instanceCredentialsLookup struct {
	once    sync.Once
	allowed bool
	known   bool
}

var instanceCredentialsLookups sync.Map

// instanceCredentialsAllowed reports whether the Platform allows instance credentials, and whether that could
// be determined. A variable so tests can replace the lookup.
var instanceCredentialsAllowed = func(ctx context.Context, client *sdk.Seqera) (allowed bool, known bool) {
	v, _ := instanceCredentialsLookups.LoadOrStore(client, &instanceCredentialsLookup{})
	lookup := v.(*instanceCredentialsLookup)
	lookup.once.Do(func() {
		lookup.allowed, lookup.known = fetchInstanceCredentialsAllowed(ctx, client)
	})
	return lookup.allowed, lookup.known
}

func fetchInstanceCredentialsAllowed(ctx context.Context, client *sdk.Seqera) (allowed bool, known bool) {
	res, err := client.ServiceInfo.Info(ctx)
	if err != nil || res == nil || res.ServiceInfoResponse == nil || res.ServiceInfoResponse.GetServiceInfo() == nil {
		tflog.Warn(ctx, "Unable to read the Platform service info; skipping plan-time AWS credential mode checks", map[string]interface{}{
			"error": fmt.Sprint(err),
		})
		return false, false
	}
	value := res.ServiceInfoResponse.GetServiceInfo().GetAllowInstanceCredentials()
	if value == nil {
		tflog.Warn(ctx, "The Platform service info does not report allowInstanceCredentials; skipping plan-time AWS credential mode checks")
		return false, false
	}
	return *value, true
}
