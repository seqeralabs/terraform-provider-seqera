# State Upgrader Guide

How to write `seqera_compute_env`-style state upgraders in this provider so that
prior state upgrades **cleanly** — without spurious "unsupported attribute"
errors when the generated schema drops fields.

**This is the default pattern.** Any new upgrader (or edit to an existing one)
should follow it. The reference implementation is `internal/stateupgraders/computeenv_v0.go`,
`computeenv_v1.go`, and the injector `internal/provider/computeenv_upgrade_schema.go`.

## TL;DR

1. **You rarely need an upgrader at all.** Speakeasy: *"adding/removing attributes
   doesn't require versioning"* (`x-speakeasy-entity-version`). Only bump the
   version for a **breaking type change** or a **rename** — a value transform the
   framework can't do on its own.
2. When you do bump the version, the upgrader must return state that decodes
   **strictly** against the current schema. Do **not** hand-strip removed
   attributes one by one — that is whack-a-mole and always incomplete.
3. Instead: do only the genuine **value transforms** on the raw JSON, then
   **re-decode leniently against the current schema** so every removed attribute
   is dropped automatically (at any nesting depth, now and on future regens).

## Why this is subtle (the mechanism)

`terraform-plugin-framework` (`internal/fwserver/server_upgraderesourcestate.go`)
has three upgrade code paths, and they treat unknown attributes **differently**:

| Path | How prior state is unmarshalled | Unknown attributes |
|------|--------------------------------|--------------------|
| stored version == current (no migration) | `UnmarshalWithOpts(..., IgnoreUndefinedAttributes: true)` | **silently dropped** |
| `StateUpgrader.PriorSchema` set | lenient unmarshal → typed `req.State`; you re-map onto current schema | **naturally absent** |
| upgrader returns `resp.DynamicValue` (raw JSON) | framework calls `DynamicValue.Unmarshal(currentType)` — **no** `IgnoreUndefinedAttributes` | **strict → hard error** |

The raw-JSON path is the "advanced" one, and it is what a hand-written upgrader
naturally uses. Its output is re-validated **strictly**, so any attribute left in
the returned state that no longer exists in the schema fails the whole upgrade:

```
Error: Unable to Upgrade Resource State
After attempting a resource state upgrade to version N, the provider returned
state data that was not compatible with the current schema.
AttributeName("compute_env").AttributeName("config").AttributeName("aws_cloud").AttributeName("enable_fusion"):
unsupported attribute "enable_fusion"
```

Two consequences:

- **Removing a field from the schema is legitimate** — the framework even has a
  built-in mechanism (`IgnoreUndefinedAttributes`) to drop it. It just doesn't
  apply that leniency to the DynamicValue output path.
- The failure is path-specific: it trips on the **first** unknown attribute it
  reaches, so fixing one (e.g. `deleted`) can reveal the next (`enable_fusion`).
  Enumerating removed attributes never converges as the schema drifts.

Upstream best practice is `StateUpgrader.PriorSchema` (see [HashiCorp: State
Upgrade](https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade)).
We can't set it here — see the next section — so we reproduce its effect.

## Why we can't just set `PriorSchema`

`PriorSchema` is a field on the `resource.StateUpgrader` **registration**, which
Speakeasy **generates** into `*_resource.go` from `x-speakeasy-entity-version`:

```go
// generated, NOT in .genignore — regenerated on every `speakeasy run`
func (r *ComputeEnvResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: stateupgraders.ComputeenvStateUpgraderV0},
		1: {StateUpgrader: stateupgraders.ComputeenvStateUpgraderV1},
	}
}
```

Speakeasy emits **no** `PriorSchema` and offers no extension to add one. Editing
the generated map would be overwritten on the next generation, and adding the
whole resource file to `.genignore` would freeze it (blocking all future schema
regens). So we keep the generated registration as-is and reproduce the lenient
decode **inside the upgrader function**, using the current schema type injected
from the provider package.

## The default pattern

### 1. Injector (provider package, `.genignore`-protected)

The upgrader functions live in `internal/stateupgraders` (the generated code
references them by name), but they need the current schema and can't import the
provider package (import cycle). Inject the schema type at `init()`:

```go
// internal/provider/computeenv_upgrade_schema.go  (add to .genignore)
package provider

// One init() registers the schema type of every resource that implements state
// upgrades — no per-resource wiring. Iterating Resources() keeps this in sync as
// resources are added.
func init() {
	ctx := context.Background()
	p := &SeqeraProvider{}
	var providerMeta provider.MetadataResponse
	p.Metadata(ctx, provider.MetadataRequest{}, &providerMeta)

	for _, newResource := range p.Resources(ctx) {
		res := newResource()
		if _, ok := res.(resource.ResourceWithUpgradeState); !ok {
			continue // only resources with upgraders need their schema registered
		}
		var meta resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: providerMeta.TypeName}, &meta)
		var schemaResp resource.SchemaResponse
		res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			continue
		}
		stateupgraders.RegisterSchemaType(meta.TypeName, schemaResp.Schema.Type().TerraformType(ctx))
	}
}
```

### 2. Shared finalizer + helper (stateupgraders/upgrade.go)

A per-resource-type registry, plus `upgradeToCurrentSchema` — the one entry point
every upgrader calls:

```go
var schemaTypes = map[string]tftypes.Type{} // keyed by resource type name

func RegisterSchemaType(resourceTypeName string, t tftypes.Type) { schemaTypes[resourceTypeName] = t }

// upgradeToCurrentSchema is the default state-upgrade implementation. It applies
// an optional in-place value transform, then re-decodes against the current
// schema, dropping (via IgnoreUndefinedAttributes, recursively) any attribute the
// schema no longer defines.
func upgradeToCurrentSchema(resourceTypeName string, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse, transform func(map[string]interface{})) {
	var rawState map[string]interface{}
	json.Unmarshal(req.RawState.JSON, &rawState) // + error handling
	if transform != nil {
		transform(rawState)
	}
	// marshal → RawState.UnmarshalWithOpts(schemaTypes[name], {IgnoreUndefinedAttributes:true})
	// → tfprotov6.NewDynamicValue → resp.DynamicValue
}
```

### 3. Each upgrader is a one-liner: type name + optional transform

```go
// No value transform — pure passthrough that drops removed attributes:
func AwscredentialStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_aws_credential", req, resp, nil)
}

// With a value transform (rename / derived value):
func ComputeenvStateUpgraderV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeToCurrentSchema("seqera_compute_env", req, resp, func(rawState map[string]interface{}) {
		if ce, ok := rawState["compute_env"].(map[string]interface{}); ok {
			applyComputeEnvV2Migrations(ce) // e.g. azure_batch string -> bool derive
		}
	})
}
```

Rules of thumb for the transform step:

- **Only encode changes that derive a new value from an old one** — renames
  (copy old key → new key before finalizing, or the value is dropped), and
  type/shape changes where the new field must be computed from the old.
- **Never `delete()` a removed attribute** — the finalizer handles removals. If
  the only change is a removal, the transform step is empty.
- The framework does **not** chain upgraders. Each registered version migrates
  **directly to the current schema**, so a `vN` upgrader must apply every
  transform from `vN` through to current (share helpers across versions).

## Testing

Put upgrader tests in the **provider package** (`internal/provider`), not
`internal/stateupgraders` — the schema injection only happens there, and testing
against the **real** schema is the whole point (synthetic fixtures miss removed
fields). Mirror the framework's strict check:

```go
value, err := resp.DynamicValue.Unmarshal(schemaType) // strict — the exact framework check
// err != nil here is the "unsupported attribute" regression.
```

Cover: (a) a populated fixture carrying attributes removed in the current schema
decodes cleanly, (b) surviving attributes are preserved, (c) each value transform
produces the right result, (d) explicitly-set values aren't clobbered.

**Also verify end-to-end for any real upgrade.** Build the old tagged provider in
a worktree, `apply` a real resource to produce genuine old-schema state, then
`plan`/`apply` with the current build and confirm a clean plan. Unit tests with
synthetic fixtures are necessary but not sufficient — the aws-cloud gap in #228
was invisible until an e2e run with a populated config.

## Decision checklist

- [ ] Is a version bump actually needed? (rename / type change → yes; add/remove → **no**)
- [ ] Does the transform step contain **only** value derivations, no `delete()` of removed fields?
- [ ] Does the upgrader delegate to `upgradeToCurrentSchema(typeName, req, resp, transform)`?
- [ ] Does every registered version migrate **directly** to the current schema?
- [ ] Is the new upgrader file added to `.genignore`? (the shared injector auto-registers its schema — no per-resource wiring)
- [ ] Tests in the provider package, asserting via the strict `DynamicValue.Unmarshal`?
- [ ] E2E: old tag → real state → current build → clean plan?

## Reference

- Shared infrastructure: `internal/stateupgraders/upgrade.go` (registry,
  `upgradeToCurrentSchema`, shared transforms), `internal/provider/stateupgrader_schemas.go` (injector)
- Example upgraders: `internal/stateupgraders/computeenv_{v0,v1}.go` (value transforms),
  any `*_credential_v0.go` (pure passthrough)
- Tests: `internal/provider/computeenv_upgrade_decode_test.go`
- Framework internals: `terraform-plugin-framework/internal/fwserver/server_upgraderesourcestate.go`,
  `terraform-plugin-go/tftypes/value_json.go`
- Issue: [#228](https://github.com/seqeralabs/terraform-provider-seqera/issues/228)
