package stateupgraders

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// schemaTypes maps a resource type name (e.g. "seqera_compute_env") to the
// tftypes.Type of its current schema. It is populated at init() by the provider
// package via RegisterSchemaType (see internal/provider/stateupgrader_schemas.go).
//
// Why injection: the upstream best practice for dropping attributes removed from
// a schema during a state upgrade is StateUpgrader.PriorSchema, which Speakeasy
// does not emit. These upgraders must live in this package (the generated
// UpgradeState() registrations reference them by name) but they need the current
// schema to re-decode against — and this package cannot import the provider
// package to build it (import cycle). So the provider package injects the schema
// types here. See docs-internal/STATE_UPGRADER_GUIDE.md.
var schemaTypes = map[string]tftypes.Type{}

// RegisterSchemaType records a resource's current schema type for use by its
// state upgraders. It is intended to be called only during package
// initialization (before any UpgradeResourceState RPC is served).
func RegisterSchemaType(resourceTypeName string, schemaType tftypes.Type) {
	schemaTypes[resourceTypeName] = schemaType
}

// upgradeToCurrentSchema is the default state-upgrade implementation for every
// resource in this provider. It decodes prior raw state, applies an optional
// in-place value transform (renames and derived values — the only things the
// framework can't do on its own), then re-encodes the state against the current
// schema, dropping any attribute the schema no longer defines.
//
// The re-decode uses IgnoreUndefinedAttributes, which recurses to every nesting
// depth — the same behavior the framework applies on its PriorSchema/passthrough
// paths, but NOT on the DynamicValue path an upgrader returns. Handling removals
// this way (rather than deleting attributes by name) is comprehensive and
// future-proof: it drops whatever the current schema no longer has, including
// attributes removed by later schema regenerations. See
// docs-internal/STATE_UPGRADER_GUIDE.md.
func upgradeToCurrentSchema(
	resourceTypeName string,
	req resource.UpgradeStateRequest,
	resp *resource.UpgradeStateResponse,
	transform func(rawState map[string]interface{}),
) {
	var rawState map[string]interface{}
	if err := json.Unmarshal(req.RawState.JSON, &rawState); err != nil {
		resp.Diagnostics.AddError("Unable to Unmarshal Prior State", err.Error())
		return
	}

	if transform != nil {
		transform(rawState)
	}

	finalizeUpgradedState(resourceTypeName, rawState, resp)
}

// finalizeUpgradedState re-encodes a migrated raw state so it is valid under the
// current schema and stores it on the response.
func finalizeUpgradedState(resourceTypeName string, rawState map[string]interface{}, resp *resource.UpgradeStateResponse) {
	schemaType, ok := schemaTypes[resourceTypeName]
	if !ok || schemaType == nil {
		resp.Diagnostics.AddError(
			"Resource Schema Not Registered",
			"No current schema type was registered for "+resourceTypeName+" before state upgrade. "+
				"This is always a bug in the provider and should be reported to the provider developer.",
		)
		return
	}

	upgradedStateJSON, err := json.Marshal(rawState)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Marshal Upgraded State", err.Error())
		return
	}

	rawUpgraded := tfprotov6.RawState{JSON: upgradedStateJSON}
	value, err := rawUpgraded.UnmarshalWithOpts(
		schemaType,
		tfprotov6.UnmarshalOpts{
			ValueFromJSONOpts: tftypes.ValueFromJSONOpts{IgnoreUndefinedAttributes: true},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode Upgraded State",
			"After applying the state migration for "+resourceTypeName+", the state could not be decoded "+
				"against the current schema. This is always a bug in the provider and should be reported "+
				"to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	dynamicValue, err := tfprotov6.NewDynamicValue(schemaType, value)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Encode Upgraded State", err.Error())
		return
	}
	resp.DynamicValue = &dynamicValue
}

// renameNvmeStorageFlag recursively renames the misspelled `nvnme_storage_enabled`
// key to `nvme_storage_enabled` at any nesting depth under the given node. It is
// shared by the compute-env family upgraders (compute_env, aws_compute_env,
// aws_batch_ce, and action, which embeds a compute env config). The exact nesting
// of the flag is not relied upon, so the rename walks the whole subtree.
func renameNvmeStorageFlag(node map[string]interface{}) {
	if v, exists := node["nvnme_storage_enabled"]; exists {
		if _, taken := node["nvme_storage_enabled"]; !taken {
			node["nvme_storage_enabled"] = v
		}
		delete(node, "nvnme_storage_enabled")
	}

	for _, child := range node {
		switch c := child.(type) {
		case map[string]interface{}:
			renameNvmeStorageFlag(c)
		case []interface{}:
			for _, item := range c {
				if m, ok := item.(map[string]interface{}); ok {
					renameNvmeStorageFlag(m)
				}
			}
		}
	}
}

// renameComputeEnvIDToID renames the root-level `compute_env_id` attribute to
// `id`. Shared by the aws_batch_ce / aws_compute_env upgraders.
func renameComputeEnvIDToID(rawState map[string]interface{}) {
	if oldValue, exists := rawState["compute_env_id"]; exists {
		if _, taken := rawState["id"]; !taken {
			rawState["id"] = oldValue
		}
		delete(rawState, "compute_env_id")
	}
}
