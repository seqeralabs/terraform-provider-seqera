# Plan: compute-environment Delete treats HTTP 400 as success

Remediation plan for [#240](https://github.com/seqeralabs/terraform-provider-seqera/issues/240).

Status: **implemented.** Root cause confirmed by experiment, and fixed via a
response-normalising SDK hook — see [Implementation](#implementation). This
route was not in the original option list; it decouples read from delete without
pinning a single generated `*_resource.go`.

The two adjacent gaps in the same hand-written delete operations — `409` not
declared, and `force` not exposed — are fixed in the same pass, so the delete
path is now uniform across all ten CE resources. See
[Rounding out the delete path](#rounding-out-the-delete-path).

## Table of Contents
- [Symptom](#symptom)
- [Root cause](#root-cause)
- [Corrections to the issue report](#corrections-to-the-issue-report)
- [Why 400 cannot simply be dropped](#why-400-cannot-simply-be-dropped)
- [Options](#options)
- [Confirming experiment](#confirming-experiment)
- [Implementation](#implementation)
- [Rounding out the delete path](#rounding-out-the-delete-path)
- [Test plan](#test-plan)
- [Related work](#related-work)

## Symptom

`terraform destroy` on a compute environment reports `Destruction complete` and
drops the resource from state even when Platform rejected the deletion with
HTTP 400. The environment survives, Terraform stops tracking it, and the
underlying cloud resources are orphaned.

Affected: the 8 typed CE resources. Each has, in its `Delete`:

```go
switch res.StatusCode {
case 204, 400, 404:
	break
default:
	resp.Diagnostics.AddError(...)
	return
}
```

`internal/provider/{awsbatchce,awscloudce,azurebatchce,azurecloudce,gcpbatchce,gcpcloudce,managedcomputece,slurmce}_resource.go`

**Not** affected: `seqera_compute_env` and `seqera_aws_compute_env`, which
tolerate only `204, 404`. That asymmetry is what locates the cause.

## Root cause

Not a Speakeasy default, and **not** the declared `400` response. The resolved
delete operations for `AWSComputeEnv` (clean) and `AWSBatchCE` (buggy) are
byte-identical apart from `operationId` / `x-speakeasy-name-override` /
`x-speakeasy-entity-operation` — same parameters, same
`responses: [204, 400, 403]`.

The cause is `x-speakeasy-entity-missing-codes` on each typed CE's **`#read`**
operation, introduced in `e8b12f4` (#201, RC-3):

```yaml
terraform-resource: AWSBatchCE#read
# 400/404 = soft-deleted/missing CE; 403 stays a real error so auth issues remain visible.
x-speakeasy-entity-missing-codes:
  - 400
  - 404
```

The extension is declared on **only the 8 GET operations — never on a delete**
(verified by walking `x-speakeasy-entity-missing-codes` across every operation
in `.speakeasy/out.openapi.yaml`). Speakeasy nonetheless applies the code set to
the entity's delete path. The correlation is exact:

| `entity-missing-codes` present | Resources | Delete tolerates |
|---|---|---|
| yes (8 overlays) | the 8 typed CEs | `204, 400, 404` |
| no (2 overlays) | `compute_env`, `aws_compute_env` | `204, 404` |

### Why the existing verification does not catch it

`internal/sdk/internal/hooks/compute_env_status_hook.go` already polls
`describe` after a delete until the environment is gone, and fails the
operation if it never disappears. It predates #240's reported commit.

It does not help here because it is an `AfterSuccess` hook that bails early:

```go
// compute_env_status_hook.go:70
if res.StatusCode != 200 && res.StatusCode != 204 { return res, nil }
```

On a `400` the polling deliberately does not run, leaving the resource's
`case 204, 400, 404` as the only thing between a rejection and an error. The two
mechanisms interlock to produce the bug.

**Consequence for the fix: reclassifying `400` is sufficient.** The
verification machinery already exists and works on the `204` path; no new
confirmation read is needed.

## Corrections to the issue report

Worth recording, because both change the scope of the work.

1. **"The provider never verifies the outcome"** — it does, via the hook above.
   The claim is what makes the reported fix look bigger than it is.
2. **The repro is partly wrong.** "A CE with active jobs" does reproduce it
   (`ActiveJobsComputeEnvValidatorImpl:44` throws `BadRequestException` → 400),
   but the two states a reader is most likely to try instead return **409**,
   which the provider already surfaces correctly:
   - `CREATING` → `ConflictException` (`ComputeEnvController:381`)
   - already `DELETING` → `ConflictException` (`ComputeEnvController:384`)

So the blast radius is the 400 class only: active jobs, unknown CE,
force-delete on a non-terminal status, and the `GlobalErrorController`
catch-all.

3. The report says "all eight compute-environment resources". There are **ten**;
   the other two are correct.

## Why 400 cannot simply be dropped

The read path genuinely needs `400`. Platform returns 400 — not 404 — for a
missing compute environment:

```groovy
// ComputeEnvServiceImpl.describe():119
if (!result)
    throw new BadRequestException("Unknown compute environment id: $computeEnvId")
```

Remove `400` from `entity-missing-codes` and refreshing a deleted CE errors
instead of dropping it from state, reintroducing the DELETED-status failure that
#201 was written to fix.

Read and delete therefore have to be decoupled. That is the whole difficulty.

## Options

In preference order *as originally assessed*. None was taken: see
[Implementation](#implementation) for the hook-based route that supersedes them.
Options 1 and 2 remain worth pursuing as the durable fixes — the hook is a
client-side shim over a Platform contract wart, not a replacement for correcting
it.

### 1. Ask Speakeasy to scope the extension (preferred first step)

The propagation from `#read` to `#delete` looks unintended. If
`x-speakeasy-entity-missing-codes` can be made operation-scoped — or if there is
a delete-specific equivalent — the fix is a one-line overlay change per file
with no generated code pinned by hand.

Cost: an upstream round trip. Do this before building a workaround.

### 2. Fix the Platform contract

Have `describe` return 404 for an unknown compute environment. Then
`entity-missing-codes: [404]` suffices, the delete tolerance disappears on its
own, and the provider needs no workaround at all.

Cleanest outcome, but it is an API contract change affecting every client, so it
needs Platform buy-in and a deprecation story.

### 3. Post-generation patch

Patch the 8 delete blocks to `case 204, 404` and add each file to `.genignore`.

Works today and is entirely within this repo. Costs 8 generated files pinned by
hand — a standing maintenance burden, and `.genignore` on a whole
`*_resource.go` freezes everything else in that file too, which is the real
objection.

**Rejected:** discriminating on the response *message* rather than the status
code, as the issue suggests. It is not expressible in an overlay and would
require option 3 anyway.

## Confirming experiment

Run this before committing to a route. It converts the propagation claim from
strong correlation into proof:

1. Remove `400` from `x-speakeasy-entity-missing-codes` in **one** overlay, e.g.
   `overlays/compute-env-slurm.yaml`.
2. `speakeasy run --skip-versioning`
3. Expect `internal/provider/slurmce_resource.go` `Delete` to become
   `case 204, 404`, and its `Read` to lose the `res.StatusCode == 400` branch.
4. Revert.

**Run 2026-08-25, and step 3 held exactly.** `slurmce_resource.go` was the only
file to change; its `Read` became `if res.StatusCode == 404` and its `Delete`
became `case 204, 404`. The propagation from `#read` to `#delete` is therefore
confirmed, not merely correlated.

## Implementation

Shipped as **option 4**, which the original list missed: normalise the response
at the SDK hook layer, so the missing-codes set never needs `400` in the first
place.

`internal/sdk/internal/hooks/compute_env_missing_hook.go` (`ComputeEnvMissingHook`)
rewrites Platform's "missing compute environment" `400` into a `404` on CE
describe and delete operations only, and only when the body carries one of the
two messages that actually mean *not found*:

| Endpoint | Message |
|---|---|
| `ComputeEnvServiceImpl.describe()` | `Unknown compute environment id: $id` |
| `ComputeEnvController.deleteComputeEnvironment()` | `Unknown computeEnv: $id` |

Note the two differ in wording; both are matched, case-insensitively.

With that in place, `x-speakeasy-entity-missing-codes` drops to `[404]` across
all 8 overlays. The propagation to `#delete` is then harmless — it propagates a
code set that no longer contains `400` — so every one of the ten CE resources
lands on `case 204, 404`, with no generated file pinned in `.genignore`.

Scoping is by operation ID: prefix `Describe`/`Delete` **and** suffix
`CE`/`ComputeEnv`. That matches exactly the 20 CE describe/delete operations and
nothing else — `Create*`, `Update*`, `Enable*`, `Disable*` and `Validate*` are
deliberately excluded, since their 400s are genuine validation failures.

Net effect on each of the ten resources:

| Platform response | Before | After |
|---|---|---|
| `400` unknown CE, on delete | silent success | tolerated as 404 (correct — it is gone) |
| `400` active jobs / force-delete misuse | **silent success** | **error, with message** |
| `400` unknown CE, on read | dropped from state | dropped from state (unchanged) |
| `400` other, on read | dropped from state | error (correct) |
| `403`, `409` | unchanged | unchanged |

### Files

- `internal/sdk/internal/hooks/compute_env_missing_hook.go` — the hook
- `internal/sdk/internal/hooks/compute_env_missing_hook_test.go` — unit tests
- `internal/sdk/computeenv_missing_integration_test.go` — end-to-end tests
  driving a real SDK client against a stub Platform, since the bug was an
  interlock between hook, SDK switch and resource tolerance rather than a fault
  in any single layer
- `internal/sdk/internal/hooks/registration.go` — registration
- `.genignore` — protects the three hand-written files above
- `overlays/compute-env-{aws-batch,aws-cloud,azure-batch,azure-cloud,gcp-batch,gcp-cloud,seqera-compute,slurm}.yaml`
  — `missing-codes` reduced to `[404]`, with a comment recording why `400` must
  not come back; plus the `force` parameter and `409` response
- `overlays/aws-compute-env.yaml` — `force` and `409` (it has no
  `missing-codes`, so it was never affected by #240)
- `overlays/compute-env.yaml` — stale comment corrected

### What 400-on-read is actually worth

Probed live against `api.cloud.dev-seqera.io` on 2026-08-25: describing *or*
deleting a non-existent `computeEnvId` returns **`403` with an empty body**, in
both workspace and user context. The permission checker
(`ComputeEnvBelongsTo{Workspace,User}Checker`) rejects an unresolvable id before
`describe()` is ever reached, so the `400` at `ComputeEnvServiceImpl:119` is not
reachable by that route.

This narrows the [Why 400 cannot simply be dropped](#why-400-cannot-simply-be-dropped)
section above: on the read path the code that matters in practice is `403`, and
`GenericResourceErrorHook` has always mapped `Describe*` `403` → `404`. So
losing 400-tolerance on read costs even less than assumed, and
`ComputeEnvMissingHook` is the defensive path for deployments or soft-delete
states that do surface the 400. It was kept rather than dropped precisely
because #201's regression is expensive and the hook is cheap.

## Rounding out the delete path

The #240 fix left two adjacent gaps in the same hand-written delete operations.
Both are fixed, so the CE delete path is now uniform across all ten resources.

### `409` is now declared

The 8 typed CE deletes (plus `AWSComputeEnv`) declared only `204/400/403`, so a
`409` — `CREATING`, or already `DELETING` — came back from the SDK as
`unknown status code returned: Status 409` and reached the user as a
`failure to invoke API` diagnostic. Platform's message was still included, so it
was never a silent-success bug, just an untyped one.

Adding `"409": Conflicting deletion` to the delete `responses` block makes it a
typed conflict that flows through the resource's `default` branch with a proper
diagnostic — matching what upstream already declares for `DeleteComputeEnv`.

### `force` is now exposed

`DELETE /compute-envs/{id}` gained a `force` query parameter in spec 1.198.0.
The legacy `seqera_compute_env` picked it up automatically because its delete
operation comes straight from upstream; the nine hand-written ones did not,
because their explicit `parameters:` lists predate it.

`force` is now declared on all nine, so every CE resource exposes:

```hcl
resource "seqera_slurm_ce" "example" {
  # ...
  force = true  # only for ERRORED / INVALID / DELETING environments
}
```

This matters because of the `204`-is-not-deleted behaviour below: without
`force`, deleting an `ERRORED` environment returns `204` and does nothing, so
`terraform destroy` hangs for ~5 minutes and then fails in the status hook with
`timeout waiting for compute environment (last status: ERRORED)`. `force = true`
is the only way out of that state.

Speakeasy threads it as an optional `types.Bool` attribute into the delete
request's query string, exactly as it already did for `seqera_compute_env`.

#### Why the parameter carries `example: false`

Speakeasy renders every optional attribute into the generated examples, and for a
bare boolean it picks `true`. That produced `force = true` in
`examples/resources/seqera_{aws_,}compute_env/resource.tf` (and so in the
rendered docs) — actively broken, since the API rejects force-delete of a healthy
`AVAILABLE` environment with a 400, exactly as `TestForceMisuseSurfacesAs400`
asserts. Anyone copying the example would have hit a failing `terraform destroy`.

The two obvious fixes are not equivalent:

| Approach | Example renders | Pre-existing state on upgrade |
|---|---|---|
| `Optional` only | `force = true` ❌ | `No changes` ✅ |
| `default: false` | `force = false` ✅ | `3 to change` — `+ force = false` on every CE ❌ |
| `example: false` | `force = false` ✅ | `No changes` ✅ |

A schema `default` makes the attribute `Computed` with
`booldefault.StaticBool(false)`, which shows `+ force = false` against every CE
already in state and forces an in-place `Update` on each — including `ERRORED` and
`INVALID` ones, which are precisely the environments `force` exists for. It also
adds a `default:"false"` tag to the SDK request, putting `force=false` on the wire
for every delete.

`example: false` fixes the example with none of that: the attribute stays plain
`Optional`, unset still means the parameter is omitted, and the third row above
was confirmed live (`terraform plan` against three real CEs: `No changes`).

The two generated examples are therefore left generated — no `.genignore` entry,
so they keep tracking future schema changes.

### No state upgrader needed

Adding an optional attribute does not require an `x-speakeasy-entity-version`
bump — per `docs-internal/STATE_UPGRADER_GUIDE.md`, only a breaking type change
or a rename does. Confirmed live: `terraform plan` against three existing
`seqera_aws_cloud_ce` resources reports `No changes`, so the new attribute is
inert for state written before it existed.

### Live end-to-end validation

Run 2026-08-25 against `api.cloud.dev-seqera.io`, using a real wedged
environment: `aws-batch-spot-5` (`2k4JDdeXZQxEn5ttnCM8VR`) in
`gavin-test-tf/genomics-research` — an AWS Batch **Forge** CE created
2025-11-26, `INVALID` with *"Associated credentials are invalid or expired"*.
Imported into a throwaway state, then destroyed.

| Step | Result |
|---|---|
| `terraform import` of the `INVALID` CE | succeeded; `status = "INVALID"` in state |
| `terraform destroy`, no `force` | **failed after 5m01s** — `timeout waiting for compute environment (last status: ERRORED)` |
| `force = true` in config only, destroy | **failed after 5m01s again** — see the gotcha below |
| `terraform apply` to persist `force`, then destroy | **succeeded in 3s** — `Destruction complete` |
| `describe` afterwards | `deleted: true`, `status: DELETED`, absent from the workspace listing |
| `terraform import` of the now-DELETED CE | `Cannot import non-existent remote object` — the #201 regression guard, passing live |

Note the status moved `INVALID` → `ERRORED` during the first failed attempt: the
non-force delete queues forge disposal, which fails, wedging it further. That is
the trap this parameter exists to escape — **5m01s and a failure, versus 3s and
a clean destroy.**

#### Gotcha: `force` must be applied, not just configured

Adding `force = true` to configuration and running `terraform destroy` does
nothing. Terraform hands the provider *prior state* for a delete, and a destroy
plan never applies pending config changes, so `Delete` still sees `force = null`.
The recipe is `terraform apply` first, then `terraform destroy`.

This is idiomatic Terraform rather than a provider defect — `force_destroy` on
`aws_s3_bucket` behaves identically — so it is fixed by documentation: the
attribute description on all nine resources now states the requirement. It
applies equally to the `force` that `seqera_compute_env` already had, where it
was never documented.

#### No orphaned cloud resources, in this case

Force-delete skips forge disposal, so the AWS resources an environment created
are normally left behind — the response confirmed it, returning eight
`forgedResources` against an empty `deletedResources`. Checked afterwards with
the `development` profile against account `128997144437`, all eight were
**already gone**: no Batch queues or compute environments, no launch template,
no IAM roles or instance profile.

That is also *why* this environment was unkillable. Its forged resources had
already been removed out of band and its credentials had expired, so Platform
held a stale `forgedResources` list and disposal could never complete. A
non-force delete had no path to success — which is exactly the state `force`
exists for.

### Added coverage

In `internal/sdk/computeenv_missing_integration_test.go`:

- `TestAllTypedComputeEnvDeletesType409` — every one of the ten delete
  operations returns a typed `409` rather than an SDK error.
- `TestForceIsSentAsQueryParam` — `force = true` reaches the wire as
  `force=true`.
- `TestForceOmittedWhenUnset` — unset means the parameter is absent, not
  `force=false`, so Platform keeps its default behaviour.
- `TestForceMisuseSurfacesAs400` — force against a non-terminal status is a real
  error, not swallowed as a missing CE.

The two `force` tests use a method-aware stub: a `204` delete makes
`ComputeEnvStatusHook` poll `describe` afterwards, and the stub must answer
`404` there for the hook to conclude the environment is gone.

## Test plan

Coverage as built. The first three are automated in
`internal/sdk/computeenv_missing_integration_test.go`, which drives a real SDK
client against a stub Platform — the layer the bug actually lived between.

- **Rejected delete surfaces an error.** ✅ `TestRejectedDeleteSurfacesAs400` and
  `TestAllTypedComputeEnvDeletesNormaliseMissing/*/rejected` — asserted for all
  nine delete operations, including that `ErrorResponse.Message` survives so the
  diagnostic keeps Platform's reason.
- **Idempotent destroy still works.** ✅ `TestDeleteMissingComputeEnvSurfacesAs404`
  and `.../*/missing`. Stronger than the original plan: a delete of an
  already-gone CE now converges *without* relying on refresh having run first.
- **409 paths unchanged.** ✅ `TestConflictingDeleteSurfacesAs409` and
  `TestAllTypedComputeEnvDeletesType409` — every delete operation returns a typed
  409 rather than the SDK's "unknown status code returned".
- **Read of a deleted CE still converges.** ✅ `TestDescribeMissingComputeEnvSurfacesAs404`,
  plus `TestDescribeOtherBadRequestStays400` for the converse — a read rejected
  for some *other* reason must not silently drop a live resource from state.
- **Hook does not misfire on healthy reads.** ✅ Live `terraform plan
  -refresh-only` against `api.cloud.dev-seqera.io` with three real
  `seqera_aws_cloud_ce` resources in state: all three refreshed, none removed,
  only `last_used` drift.
- **Unit coverage.** ✅ Ten cases in
  `internal/sdk/internal/hooks/compute_env_missing_hook_test.go` covering
  operation scoping, message matching, body preservation, oversized bodies, nil
  responses, and the untouched status codes.

Also run: `go build`, `go test ./...`, `gofmt`, `go vet`, and staticcheck with
the same flags `speakeasy run` uses — all clean.

Note on staticcheck: run it the way `speakeasy run` does, via
`go run honnef.co/go/tools/cmd/staticcheck@v0.7.0`, so it compiles against the
active toolchain. A `staticcheck` binary already on `$PATH` may be too old to
build this module at all — `~/go/bin/staticcheck` 0.6.1 (built with go1.25.5)
fails with *"module requires at least go1.25.10, but Staticcheck was built with
go1.25.5"* and reports nothing. Also beware `staticcheck ... | head`, where `$?`
is `head`'s status, not staticcheck's; redirect to a file and check the exit code
directly.

### Found while testing: typed CE deletes declared neither 409 nor `force`

Both fixed in the same pass — see [Rounding out the delete path](#rounding-out-the-delete-path).

## Related work

- ✅ **`409` not declared on typed CE deletes** — fixed here.
- ✅ **`force=true` on typed CE deletes** — fixed here. Note the plan previously
  recorded this as "patch staged on branch `feat/ce-force-delete`"; that branch
  does not exist locally or on the remote, so it was written from scratch.
- **`204` does not mean deleted.** For Forge environments,
  `ComputeEnvDeleteServiceImpl` sets status to `DELETING`, queues async forge
  disposal and returns; the controller answers `204` regardless. The status hook
  covers this by polling, and that is why a non-force delete of an `ERRORED`
  environment returns `204` and changes nothing. Confirmed live 2026-08-25.
- ✅ Without `force`, deleting an `ERRORED` CE returns `204` and does nothing, so
  `terraform destroy` hangs ~5 min then fails in the status hook with
  `timeout waiting for compute environment (last status: ERRORED)`. Fixing #240
  did not address that path — `force = true` now does, on all ten resources.
  Verified live against a genuinely wedged environment; see
  [Live end-to-end validation](#live-end-to-end-validation).
