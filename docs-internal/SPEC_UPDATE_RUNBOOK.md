# Updating the API Specification

This guide explains how to update the Seqera API OpenAPI specification used by
the provider, review the generated changes, and verify the result.

Fetching and regenerating are mostly mechanical. The important review happens
after generation: for each new attribute, decide whether it belongs in the
Terraform schema and whether it needs validation or an overlay.

## Contents

- [When to run this](#when-to-run-this)
- [Fast path](#fast-path)
- [1. Vendor the new spec](#1-vendor-the-new-spec)
- [2. Review the spec diff](#2-review-the-spec-diff)
- [3. Regenerate](#3-regenerate)
- [4. Review the generated schema](#4-review-the-generated-schema)
- [5. Verify](#5-verify)
- [6. Commit and PR](#6-commit-and-pr)
- [Common issues](#common-issues)

## When to run this

Run this when the platform publishes API changes that should be available in
the provider, or before a provider release. The specification is checked into
the repository so every update can be reviewed.

You will need the `speakeasy` CLI, `npx`, and the Go toolchain. If you have
access to a checkout of the platform backend or frontend, those codebases can
help when the OpenAPI description does not explain a field's behaviour.

First check whether the vendored version is behind the published version:

```bash
make spec-version      # vendored vs. latest published
```

## Fast path

`make update-spec-online` automates fetching, formatting, regeneration, and the
build, then prints the generated documentation diff. It changes your working
tree but does not commit anything. It does not pause for the spec-diff review
in step 2, and it stops before schema triage because the generated changes
still need a human review.

```bash
make update-spec-online
```

Then go straight to [step 4](#4-review-the-generated-schema).

The sections below describe the same process in detail. Use them when you need
to inspect the specification before regeneration or troubleshoot a failed
update. Reviewing the spec diff first makes the generated changes easier to
understand.

## 1. Vendor the new spec

The provider builds from `specs/seqera-api-cloud.yaml`, which is declared as
the source input in `.speakeasy/workflow.yaml`. Replace this file with the
published specification. Provider-specific customisation belongs in
`overlays/`, so the vendored file should not contain `x-speakeasy-*`
annotations. Check that before fetching:

```bash
grep -Ec '^[[:space:]]*x-speakeasy-' specs/seqera-api-cloud.yaml   # must be 0
```

Then fetch and format the new file. `make fetch-spec` performs both checks for
hand-added annotations and leaves the existing file untouched if the download
fails:

```bash
make fetch-spec
```

The equivalent manual steps are:

```bash
curl -fsSL -o /tmp/latest.yml \
  https://cloud.seqera.io/openapi/seqera-api-latest-flattened.yml

/bin/cp -f /tmp/latest.yml specs/seqera-api-cloud.yaml
make format-spec
```

`make format-spec` re-sorts the spec using
`specs/.openapi-format-sort.yaml`. This keeps the diff focused on actual API
changes. Without it, the different ordering used by the upstream file can
produce a large, mostly mechanical diff. If the `make` target is unavailable,
run it directly:

```bash
npx openapi-format specs/seqera-api-cloud.yaml -o specs/seqera-api-cloud.yaml \
  --sortComponentsProps -s specs/.openapi-format-sort.yaml --yamlQuoteStyle single
```

As a quick check, confirm that the file starts with `openapi: 3.0.1` and that
the diff is roughly the size expected for the release:

```bash
head -8 specs/seqera-api-cloud.yaml
git diff --stat specs/seqera-api-cloud.yaml
```

## 2. Review the spec diff

Read the spec diff before regenerating. Look for new, removed, or renamed
paths, schemas, and properties.

New paths and schemas:

```bash
git diff -U0 specs/seqera-api-cloud.yaml \
  | grep -E '^\+ {4}[A-Za-z].*:$|^\+ {2}/'
```

Removals and renames deserve particular attention because they can affect
existing configurations:

```bash
git diff specs/seqera-api-cloud.yaml \
  | grep -E '^-[^-]' | grep -v '^-[[:space:]]'
```

For anything renamed or removed, check whether an overlay references the old
name. An overlay targeting a path that no longer exists may be skipped without
an error:

```bash
grep -rn 'oldFieldName\|OldSchemaName' overlays/
```

## 3. Regenerate

```bash
speakeasy run --skip-versioning
go build -o terraform-provider-seqera
```

If generation fails with duplicate type declarations, check for stale
untracked files under `internal/` from a previous partial run. Review
`git status`, remove only the generated files that do not belong in the working
tree, and retry.

## 4. Review the generated schema

`docs/resources/*.md` shows what reached the Terraform schema. Review this
diff after every regeneration:

```bash
make docs-diff         # changed files, plus the added attributes
git diff -U5 -- docs/  # the full picture
```

For each added attribute, first check what it means in the platform backend.
The OpenAPI description may be incomplete, and some important constraints are
not represented in the schema. If the platform codebase is available, search
it from the repository root:

```bash
grep -rn 'fieldName' --include='*.groovy' --include='*.java' . | grep -v '/test/'
```

If the frontend codebase is available, check whether it uses the field there as
well. If neither codebase is available, use the API documentation and confirm
uncertain behavior with the platform team before exposing the field in
Terraform.

Choose one of the following outcomes for each attribute.

### Keep it

Keep the attribute when it represents user-configurable settings. The
generator may already have added it. Confirm that it is `Optional` unless the
API genuinely requires it, and that `ForceNew` matches whether the platform
supports updating it in place.

### Ignore it

Ignore the attribute when it is not stable Terraform configuration. A useful
test is whether two users could see different values for the same resource.
Per-user state, server timestamps, and runtime status can all create perpetual
diffs and generally do not belong in Terraform state.

Prefer `x-speakeasy-terraform-ignore` over `remove: true`. This keeps the field
in the SDK model while removing it from the Terraform schema, so API responses
can still be decoded:

```yaml
- target: $.components.schemas.SomeDto.properties.starred
  update:
    x-speakeasy-param-readonly: true
    x-speakeasy-terraform-ignore: true
```

### Validate it

Add a validator when the API documents a constraint that is not enforced by the
generated schema. Plan-time validation gives users feedback before apply.

With the current Speakeasy version, `minimum` and `maximum` in the spec do not
generate validators. For example, `bid_percentage` in
`overlays/compute-env.yaml` has both constraints but only gets plan modifiers.
Add a validator explicitly:

1. Write it in `internal/validators/<type>validators/<name>.go`, following the
   patterns in [OVERLAY_GUIDE.md](./OVERLAY_GUIDE.md#custom-validators).
2. Wire it with `x-speakeasy-plan-validators: YourValidatorName`. The name
   resolves to `custom_<type>validators.YourValidatorName()`, so an
   `Int32Attribute` validator must live in `int32validators`.
3. **Add the file to `.genignore`**, or the next `speakeasy run` deletes it.

Apply the validator to the shared schema where possible. For example, a
validator on `SchedConfig` applies to every compute environment that embeds it.

Cross-field rules belong here too. If a field is only valid under a particular
condition, read the sibling with
`req.Path.ParentPath().AtName("other_field")`. Treat a null sibling as the API
default when appropriate, and report an error only for an explicit conflict.

### Constrain it

Use this when an enum has gained a value that is not currently usable. Prefer a
description over restricting the enum when possible; see
[forward compatibility](#constraining-an-enum-breaks-forward-compatibility)
below.

### Sync a description

Use this when an overlay contains a description that is more specific than the
specification. If upstream changes the field's meaning, update the overlay as
well. Search for the existing wording:

```bash
grep -rn 'old phrasing' overlays/
```

### Endpoints without resources

New paths generate SDK clients, but they do not generate resources or data
sources unless an overlay opts them in. That is expected. Mention intentionally
unexposed endpoints in the PR so the decision is clear.

## 5. Verify

```bash
go build -o terraform-provider-seqera && go test ./internal/...
```

Then exercise the schema with a local `dev_overrides` plan. A create-only plan
against the built binary does not touch live state or make resource API calls,
so it can be run without credentials:

```bash
mkdir -p /tmp/spec-check && cd /tmp/spec-check
cat > .terraformrc <<'EOF'
provider_installation {
  dev_overrides {
    "registry.terraform.io/seqeralabs/seqera" = "/Users/you/Code/terraform-provider-seqera"
  }
  direct {}
}
EOF
# main.tf: one resource per new attribute, plus the negative cases
TF_CLI_CONFIG_FILE=$PWD/.terraformrc terraform plan
```

Cover three things:

1. **Every new attribute set to a realistic value** — confirms that the schema
   accepts the value and uses the expected type.
2. **Each validator's negative cases** — include out-of-range values and any
   cross-field conflicts. Check the error text as well as the failure.
3. **An existing resource shape with the new optional fields unset** — checks
   that the new validation does not reject an existing configuration.

This does not catch refresh-time issues or spurious diffs on resources that are
already in state. Call out that limitation in the PR.

## 6. Commit and PR

Keep the update in one commit where practical. The commit message and PR body
should explain what changed and why. Include:

- new attributes and which resources they reached
- endpoints that generate SDK clients but deliberately no resources
- what was verified, and **what wasn't**

## Common issues

These issues are easy to miss during a spec update.

### Sibling keys next to a `$ref` are ignored

```yaml
mode:
  $ref: '#/components/schemas/SomeEnum'
  description: 'This never appears anywhere.'   # silently dropped
```

The resolver discards siblings of a `$ref`. The overlay may apply, but the
description will not appear in the generated output. Either wrap the reference
in `allOf` or inline the definition.

### Constraining an enum breaks forward compatibility

Removing a value from an enum does not just reject it as *input*. Speakeasy
folds a single-consumer shared enum into a local type whose generated
`UnmarshalJSON` **errors on any unlisted value**:

```go
default:
    return fmt.Errorf("invalid value for Mode: %v", v)
```

As a result, a resource created elsewhere with the new value can no longer be
*read*. The failure appears as an unmarshal error during refresh.

Keep the value in the enum and document its current limitations when possible.
Only restrict it when the value is unusable everywhere; doing so trades
forward-compatible reads for stricter plan-time validation.

### Multi-line descriptions break the docs list

`docs/resources/*.md` renders attribute descriptions inside a Markdown list.
A multi-line description with sub-bullets can therefore look like additional
*attributes*:

```markdown
- `mode` (String) Authentication mode:
- `keys` (default): static access key      <- looks like an attribute named keys
```

Keep attribute descriptions to a single line. Multi-line text is fine for
resource-level descriptions. Always read the generated `docs/` diff, not just
the overlay.

### Check the generated docs

An overlay can apply cleanly without producing the output you intended. Check
`git diff -- docs/` after regenerating to confirm its effect.

### Hand-written files need `.genignore`

Files under `internal/` that are not generated — such as validators, state
upgraders, and SDK hooks — must be listed in `.genignore` or the next
`speakeasy run` may delete them.

### The spec may not describe every backend rule

Ranges, mutual exclusions, and feature gates are sometimes enforced by the
backend without appearing in the spec. A value can therefore be valid in the
schema but fail at apply time. Query the platform codebase when it is available
and document constraints that cannot be enforced during planning.

## See also

- [OVERLAY_GUIDE.md](./OVERLAY_GUIDE.md) — overlay structure, field cleanup,
  validator patterns
- [SPEAKEASY_EXTENSIONS_REFERENCE.md](./SPEAKEASY_EXTENSIONS_REFERENCE.md) —
  what each `x-speakeasy-*` extension does
- [STATE_UPGRADER_GUIDE.md](./STATE_UPGRADER_GUIDE.md) — needed only when a bump
  forces a schema version change
