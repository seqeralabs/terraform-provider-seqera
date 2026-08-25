# API Spec Update Runbook

How to pull a new Seqera API OpenAPI spec into the provider, decide what each
change should mean in Terraform, and verify the result.

The mechanical part (fetch, regenerate) takes minutes. The part that needs
judgement is **step 4** — deciding, for every attribute the bump adds, whether
it belongs in the Terraform schema at all. Most of this guide is about that.

## Table of Contents
- [When to run this](#when-to-run-this)
- [Fast path](#fast-path)
- [1. Vendor the new spec](#1-vendor-the-new-spec)
- [2. Review the spec diff](#2-review-the-spec-diff)
- [3. Regenerate](#3-regenerate)
- [4. Triage the docs diff](#4-triage-the-docs-diff)
- [5. Verify](#5-verify)
- [6. Commit and PR](#6-commit-and-pr)
- [Traps](#traps)

## When to run this

When the platform has shipped API changes you want to expose, or before a
release so the provider isn't lagging. There is no automation — the spec is
vendored deliberately so that a bump is always reviewed.

Prerequisites: `speakeasy` CLI, `npx`, Go toolchain, and a checkout of the
platform backend at `~/Code/platform` (needed in step 4 to answer "what does
this field actually do?").

Check whether a bump is even needed:

```bash
make spec-version      # vendored vs. latest published
```

## Fast path

`make update-spec-online` does steps 1-3 in one go: fetch, sort, regenerate,
build, then print the generated-docs diff. It stops there on purpose and
commits nothing, because step 4 is a judgement call.

```bash
make update-spec-online
```

Then go straight to [step 4](#4-triage-the-docs-diff).

The rest of this guide is the same process broken out, for when something goes
wrong or you want to inspect the spec diff before regenerating. Read step 2 at
least once — reviewing the spec diff before generation is much easier than
reverse-engineering the result afterwards.

## 1. Vendor the new spec

The provider builds from `specs/seqera-api-cloud.yaml`, declared as the source
input in `.speakeasy/workflow.yaml`. Overwrite it wholesale — all our
customisation lives in `overlays/`, and the vendored spec carries **no**
`x-speakeasy-*` annotations of its own. Verify that before trusting a blind
overwrite:

```bash
grep -c 'x-speakeasy' specs/seqera-api-cloud.yaml   # must be 0
```

Then fetch and format. `make fetch-spec` does both, refuses to run if the
vendored spec has been hand-annotated, and leaves the spec untouched if the
download fails:

```bash
make fetch-spec
```

Equivalent by hand:

```bash
curl -fsSL -o /tmp/latest.yml \
  https://cloud.seqera.io/openapi/seqera-api-latest-flattened.yml

/bin/cp -f /tmp/latest.yml specs/seqera-api-cloud.yaml
make format-spec
```

`make format-spec` re-sorts the spec with `specs/.openapi-format-sort.yaml`.
**Do not skip it.** Upstream serves keys in a different order (`components`
first, alphabetical); without the sort you get a ~10k-line reordering diff
instead of a reviewable one. If the `make` target is unavailable, run it
directly:

```bash
npx openapi-format specs/seqera-api-cloud.yaml -o specs/seqera-api-cloud.yaml \
  --sortComponentsProps -s specs/.openapi-format-sort.yaml --yamlQuoteStyle single
```

Sanity check — the file should now start with `openapi: 3.0.1`, and the diff
should be proportionate to the release:

```bash
head -8 specs/seqera-api-cloud.yaml
git diff --stat specs/seqera-api-cloud.yaml
```

## 2. Review the spec diff

Read the spec diff *before* regenerating, so you know what to expect from the
generator rather than reverse-engineering it afterwards.

New paths and schemas:

```bash
git diff -U0 specs/seqera-api-cloud.yaml \
  | grep -E '^\+ {4}[A-Za-z].*:$|^\+ {2}/'
```

Removals and renames — the short list that matters most, since these are the
only things that can break existing configs:

```bash
git diff specs/seqera-api-cloud.yaml | grep -E '^-[^-]' | grep -v '^-\s'
```

For anything renamed or removed, check whether an overlay references the old
name. An overlay targeting a path that no longer exists **fails silently** —
the action is skipped, not errored:

```bash
grep -rn 'oldFieldName\|OldSchemaName' overlays/
```

## 3. Regenerate

```bash
speakeasy run --skip-versioning
go build -o terraform-provider-seqera
```

If generation fails with duplicate type declarations, you are probably hitting
stale untracked files under `internal/` from a previous partial run. Check
`git status` for untracked `internal/**` files, clean them, and retry.

## 4. Triage the docs diff

`docs/resources/*.md` is the readable summary of what actually reached the
Terraform surface. This diff *is* the review:

```bash
make docs-diff         # changed files, plus the added attributes
git diff -U5 -- docs/  # the full picture
```

Every added attribute needs a decision. Before deciding, look up what the field
means in the platform backend — the spec description is often thin, and
occasionally the constraint that matters isn't expressed in the schema at all:

```bash
cd ~/Code/platform
grep -rn 'fieldName' --include='*.groovy' --include='*.java' . | grep -v '/test/'
```

Also check the frontend (`tower-web/src/app/data/entity/`) — if the UI doesn't
offer a field, that's a strong signal the backend is ahead of the product and
the field isn't ready to expose.

There are five outcomes. Pick one per attribute.

### Keep it

Real, user-settable configuration. Nothing to do — the generator already did
it. Confirm it landed as `Optional` (not `Required`) unless the API genuinely
demands it, and that `ForceNew` matches whether the platform can update it
in place.

### Ignore it

The field is real but has no business in Terraform state. The clearest test:
**would two different readers see different values?** Per-user state
(who starred a studio), server-side timestamps, and runtime status all fail
that test and generate perpetual diffs.

Prefer `x-speakeasy-terraform-ignore` over `remove: true` — it keeps the field
in the SDK model and drops it only from the Terraform schema, so reads still
deserialize:

```yaml
- target: $.components.schemas.SomeDto.properties.starred
  update:
    x-speakeasy-param-readonly: true
    x-speakeasy-terraform-ignore: true
```

### Validate it

The API documents a constraint it doesn't express in the schema. Add a plan-time
validator — this provider errors at plan time rather than deferring to a
permissive backend, because a silent no-op is worse than a failed plan.

**`minimum`/`maximum` in the spec generate no validator** with the current
Speakeasy version. Compare `bid_percentage` in `overlays/compute-env.yaml`,
which has both and renders only plan modifiers. You need a real validator:

1. Write it in `internal/validators/<type>validators/<name>.go`, following the
   patterns in [OVERLAY_GUIDE.md](./OVERLAY_GUIDE.md#custom-validators).
2. Wire it with `x-speakeasy-plan-validators: YourValidatorName`. The name
   resolves to `custom_<type>validators.YourValidatorName()`, so an
   `Int32Attribute` validator must live in `int32validators`.
3. **Add the file to `.genignore`**, or the next `speakeasy run` deletes it.

Target the shared schema, not each resource — a validator on `SchedConfig`
propagates to every CE that embeds it.

Cross-field rules belong here too. If a field is only honoured under some
condition, say so: reading a sibling via
`req.Path.ParentPath().AtName("other_field")` is the standard shape. Mind the
defaults — a *null* sibling means the API default, which may well satisfy the
condition, so only error on an explicit conflicting value.

### Constrain it

The enum gained a value that isn't usable yet. Think hard before doing this,
and prefer documenting over constraining — see
[forward compatibility](#constraining-an-enum-breaks-forward-compatibility)
below.

### Sync a description

Several overlays hardcode descriptions that shadow the spec's own. When
upstream revises the wording — a new suggested enum value, a corrected default —
the overlay silently keeps the stale text. Search for the old phrasing:

```bash
grep -rn 'old phrasing' overlays/
```

### Endpoints without resources

New paths generate SDK clients but no resources or data sources unless an
overlay opts them in. That's the correct default. Note them in the PR so
reviewers know the omission was deliberate, not missed.

## 5. Verify

```bash
go build -o terraform-provider-seqera && go test ./internal/...
```

Then exercise the schema for real. A throwaway directory with `dev_overrides`
plans against the built binary without touching any live state — no backend, no
credentials, and a create-only plan makes no API calls for the resources
themselves:

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

1. **Every new attribute set to a realistic value** — proves the schema accepts
   it and the type is right.
2. **Each validator's negative cases** — out of range, and any cross-field
   conflict. Assert on the error text, not just that it failed.
3. **A resource shaped like one that already exists in the wild**, with the new
   optional fields *unset*. This is the regression case: a new validator that
   misfires on an existing config is the most likely way to break users.

Note what this can't reach: a create-only plan never refreshes, so it won't
catch a field that produces a spurious diff on a resource already in state.
Say so in the PR rather than implying full coverage.

## 6. Commit and PR

One commit. The message should say what the bump added and, for every
adjustment, *why* — the reasoning is the part nobody can reconstruct from the
diff later. Same for the PR body, plus:

- new attributes and which resources they reached
- endpoints that generate SDK clients but deliberately no resources
- what was verified, and **what wasn't**

## Traps

Each of these has cost real debugging time.

### Sibling keys next to a `$ref` are ignored

```yaml
mode:
  $ref: '#/components/schemas/SomeEnum'
  description: 'This never appears anywhere.'   # silently dropped
```

The resolver discards siblings of a `$ref`. A description written this way
never reaches the generated docs, and it fails quietly — the overlay applies,
the text just evaporates. Either wrap in `allOf`, or inline the definition.

### Constraining an enum breaks forward compatibility

Removing a value from an enum doesn't just reject it as *input*. Speakeasy
folds a single-consumer shared enum into a local type whose generated
`UnmarshalJSON` **errors on any unlisted value**:

```go
default:
    return fmt.Errorf("invalid value for Mode: %v", v)
```

So a resource created elsewhere with the new value can no longer be *read* —
the failure shows up as an unmarshal error on refresh, not a clean rejection.

Default to keeping the value in the enum and documenting its limits in the
description. Reserve constraining for a value that is genuinely unusable
everywhere, and know you're trading read tolerance for plan-time strictness.

### Multi-line descriptions break the docs list

`docs/resources/*.md` splices attribute descriptions into a Markdown list, so a
multi-line description with sub-bullets renders as sibling *attributes*:

```markdown
- `mode` (String) Authentication mode:
- `keys` (default): static access key      <- looks like an attribute named keys
```

Keep attribute descriptions to a single line. Multi-line is fine for
schema-level (resource) descriptions. Always read the rendered `docs/` diff, not
just the overlay.

### The generated docs are the only place the truth shows up

An overlay can apply cleanly and still do nothing useful (see the `$ref` trap).
`git diff -- docs/` after regenerating is the check that the overlay had the
effect you intended.

### Hand-written files need `.genignore`

Anything you write under `internal/` — validators, state upgraders, SDK hooks —
is deleted by the next `speakeasy run` unless listed in `.genignore`.

### The spec is not the authority on semantics

Ranges, mutual exclusions, and feature gating are frequently absent from the
spec but enforced in the backend. A field can be fully valid per the schema and
still fail at apply time because it sits behind a feature flag or is rejected
for a particular platform. `~/Code/platform` is the source of truth; when a
constraint can't be checked at plan time, put it in the description so users
aren't surprised.

## See also

- [OVERLAY_GUIDE.md](./OVERLAY_GUIDE.md) — overlay structure, field cleanup,
  validator patterns
- [SPEAKEASY_EXTENSIONS_REFERENCE.md](./SPEAKEASY_EXTENSIONS_REFERENCE.md) —
  what each `x-speakeasy-*` extension does
- [STATE_UPGRADER_GUIDE.md](./STATE_UPGRADER_GUIDE.md) — needed only when a bump
  forces a schema version change
