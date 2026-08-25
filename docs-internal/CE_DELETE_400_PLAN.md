# Plan: compute-environment Delete treats HTTP 400 as success

Remediation plan for [#240](https://github.com/seqeralabs/terraform-provider-seqera/issues/240).

Status: **investigated, not started.** Root cause confirmed; the fix route
needs one decision (see [Options](#options)).

## Table of Contents
- [Symptom](#symptom)
- [Root cause](#root-cause)
- [Corrections to the issue report](#corrections-to-the-issue-report)
- [Why 400 cannot simply be dropped](#why-400-cannot-simply-be-dropped)
- [Options](#options)
- [Confirming experiment](#confirming-experiment)
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

In preference order.

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

If step 3 holds, the mechanism is confirmed and option 1 is precisely the right
question to put to Speakeasy.

## Test plan

Whichever route is taken:

- **Rejected delete surfaces an error.** Point a CE at a workspace where the
  delete will 400 (a CE with an active run is the reliable trigger), then
  `terraform destroy` and assert it fails with the response body in the
  diagnostic, and that the resource stays in state.
- **Idempotent destroy still works.** Delete a CE out of band, then
  `terraform destroy`. Refresh should drop it from state via the read path
  before delete is ever called — confirm that is what happens, since it is the
  argument for why losing 400-tolerance on delete costs little.
- **409 paths unchanged.** Destroy a CE in `CREATING`; assert the existing
  conflict error is still surfaced.
- **Read of a deleted CE still converges.** Delete out of band, `terraform plan`,
  assert the resource is removed from state and not errored — the regression
  guard for #201.

## Related work

- **`force=true` on typed CE deletes.** `DELETE /compute-envs/{id}` gained a
  `force` parameter (present as of spec 1.198.0), and `seqera_compute_env`
  already exposes it (`computeenv_resource.go:5858`). The 8 typed resources do
  not, because their delete operations are hand-written in the overlays with
  explicit `parameters:` lists that predate the parameter. Patch staged on
  branch `feat/ce-force-delete` (9 overlays, not yet regenerated).
- **`204` does not mean deleted.** For Forge environments,
  `ComputeEnvDeleteServiceImpl` sets status to `DELETING`, queues async forge
  disposal and returns; the controller answers `204` regardless. The status hook
  covers this by polling, and that is why a non-force delete of an `ERRORED`
  environment returns `204` and changes nothing. Confirmed live 2026-08-25.
- Without `force`, deleting an `ERRORED` CE returns `204` and does nothing, so
  `terraform destroy` hangs ~5 min then fails in the status hook with
  `timeout waiting for compute environment (last status: ERRORED)`. Fixing #240
  does not address that path; `force` does.
