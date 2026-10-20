# Workspace Lineage Settings Resource Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `seqera_workspace_lineage_settings` Terraform resource backed by the `GET/PUT/DELETE /orgs/{orgId}/workspaces/{workspaceId}/settings/lineage` endpoints.

**Architecture:** Speakeasy generates the resource from a new overlay that binds the `PUT` to create+update, `GET` to read and `DELETE` to delete. Hand-written, `.genignore`-protected pieces cover what the generator cannot: two SDK hooks (poll AUTO provisioning after `PUT`; turn an "unconfigured" 200 on `GET` into a 404 so Terraform drops the resource) and one plan validator for the MANUAL/AUTO field rules.

**Tech Stack:** Go, terraform-plugin-framework, Speakeasy 1.802.0 (overlays in `overlays/`, vendored spec `specs/seqera-api-cloud.yaml`).

**Spec:** `docs/superpowers/specs/2026-10-20-workspace-lineage-settings-design.md`

## Global Constraints

- Resource name `seqera_workspace_lineage_settings` (Speakeasy entity `WorkspaceLineageSettings`).
- Do not hand-edit generated files. Every hand-written file under `internal/` is added to `.genignore` in the same commit that creates it.
- Attribute descriptions are single-line strings (docs list rendering).
- Provisioning wait: 10 s poll interval, 10 min overall timeout, no `timeouts` block.
- `queue_arn` and `store_uri` are not exposed in the schema (`x-speakeasy-terraform-ignore: true`).
- `webhook_url` is `Sensitive`.
- `provisioning_mode` is Optional+Computed with default `AUTO`.
- `store_name`/`topic_arn` are Optional+Computed, with no `UseStateForUnknown`; validators read the configured value only, never state.
- Annotations live in a new overlay (`overlays/workspace-lineage-settings.yaml`) registered in `.speakeasy/workflow.yaml`. `CLAUDE.md` mentions `schemas/seqera-final.yaml`; that file does not exist in this repo, so do not create it.
- Verification is `go build`, `go test ./internal/...` and local `terraform plan` (no live backend).

## Review Focus

- `GET` on an unconfigured workspace (200, null `credentialsId`) must remove the resource from state on read, not error or yield a half-empty resource. Pinned in Task 3.
- `PUT` succeeding while AUTO ends `ERRORED` must fail the apply with the API `message`, not report success. Pinned in Task 2.
- Polling must survive transient 5xx/429 and network errors, and honour context cancellation. Pinned in Task 2.
- A user with no `provisioning_mode` in config must be treated as AUTO by the validator (null == AUTO). Pinned in Task 4.
- Unknown values (e.g. `topic_arn` coming from another resource) must not raise validation errors. Pinned in Task 4.

---

## File Structure

| File | Responsibility |
|---|---|
| `overlays/workspace-lineage-settings.yaml` (create) | Entity and operation bindings, field config, descriptions |
| `.speakeasy/workflow.yaml` (modify) | Register the overlay |
| `internal/sdk/internal/hooks/lineage_settings_status_hook.go` (create) | Poll after `PUT` until status leaves `PENDING` |
| `internal/sdk/internal/hooks/lineage_settings_not_configured_hook.go` (create) | Map unconfigured `GET` 200 to 404 |
| `internal/sdk/internal/hooks/registration.go` (modify) | Register both hooks |
| `internal/validators/stringvalidators/lineage_manual_field_validator.go` (create) | MANUAL-required / AUTO-forbidden check for `store_name` and `topic_arn` |
| `examples/resources/seqera_workspace_lineage_settings/resource.tf` (create) | AUTO and MANUAL examples |
| `.genignore` (modify) | Protect every hand-written file above |

Tests sit beside each hand-written file (`*_test.go`).

---

### Task 1: Overlay and generation spike

Decides whether the generated route works (spec "Left to the builder"). The rest of the plan assumes it does; if it does not, stop and bring the finding to the user before continuing.

**Files:**
- Create: `overlays/workspace-lineage-settings.yaml`
- Modify: `.speakeasy/workflow.yaml` (add the overlay after `overlays/workspaces.yaml`)
- Generated (do not hand-edit): `internal/provider/workspacelineagesettings_resource.go` and `_resource_sdk.go`, `docs/resources/workspace_lineage_settings.md`

**Interfaces:**
- Produces: resource `seqera_workspace_lineage_settings` with attributes `org_id`, `workspace_id` (required, `RequiresReplace`), `credentials_id`, `region` (required), `provisioning_mode`, `lineage_default`, `store_name`, `topic_arn` (inputs), and `status`, `message`, `event_delivery`, `webhook_url` (Sensitive), `provider`, `provisioned` (read-only). Attribute `store_name` carries `x-speakeasy-plan-validators: LineageManualFieldValidator`, and so does `topic_arn` (the validator is created in Task 4).
- Import id is the repo-standard JSON object string produced by the generator (`{"org_id": 0, "workspace_id": 0}`), not `<org_id>/<workspace_id>`. This deviates from the design doc; flag it in the PR.

- [ ] **Step 1: Write the overlay.** Follow `overlays/orgs.yaml` and `overlays/primary-compute-env.yaml`. Bind `PUT` to `WorkspaceLineageSettings#create` and `#update`, `GET` to `#read`, `DELETE` to `#delete`. If one operation cannot carry two bindings, use Speakeasy's documented syntax for it. Set `x-speakeasy-terraform-ignore: true` on `queueArn` and `storeUri`, mark `webhookUrl` sensitive, default `provisioningMode` to `AUTO`, and mark `storeName` and `topicArn` Optional+Computed.
- [ ] **Step 2: Register the overlay in `.speakeasy/workflow.yaml`, run `speakeasy run --skip-versioning`.** Expected: generation succeeds. The new resource is wired in `internal/provider/provider.go`.
- [ ] **Step 3: Run `go build ./... && go test ./internal/...`.** Expected: PASS. The `LineageManualFieldValidator` reference will not compile until Task 4; for this spike, temporarily omit `x-speakeasy-plan-validators`, and add it back in Task 4.
- [ ] **Step 4: Inspect the generated Create.** Confirm it calls `UpdateLineageWorkspaceSettings` (a 204 with no body) and then refreshes state through `FindLineageWorkspaceSettings`. If it does not, record this in the PR description and stop; the fallback is a hand-written resource protected by `.genignore`, which changes the remaining plan and needs the user's go-ahead.
- [ ] **Step 5: Commit** `feat(lineage): generate seqera_workspace_lineage_settings resource`.

---

### Task 2: Provisioning status hook

**Files:**
- Create: `internal/sdk/internal/hooks/lineage_settings_status_hook.go`
- Test: `internal/sdk/internal/hooks/lineage_settings_status_hook_test.go`
- Modify: `internal/sdk/internal/hooks/registration.go`, `.genignore`

**Interfaces:**
- Produces: `type LineageSettingsStatusHook struct { InitialWait, PollInterval, RetryInterval, OverallTimeout time.Duration }`. A zero field means its default (2 s, 10 s, 1 s, 10 min). It has `AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error)`, which acts only on `hookCtx.OperationID == "UpdateLineageWorkspaceSettings"` with status 204.

- [ ] **Step 1: Write failing tests** with `httptest.Server` and tiny durations: `TestLineageSettingsStatusHook_AvailableImmediately` (GET returns `AVAILABLE`, no error), `_PendingThenAvailable` (two `PENDING` then `AVAILABLE`), `_ErroredReturnsMessage` (status `ERRORED`, `message` "boom": the error contains "boom"), `_TimesOutWhilePending` (error mentions the last status `PENDING`), `_RetriesOn503ThenSucceeds`, `_IgnoresOtherOperations` (a different `OperationID` makes no HTTP call), `_ContextCancelled`. The test request carries `Authorization: Bearer x` and the path `/orgs/7/workspaces/9/settings/lineage`.
- [ ] **Step 2: Run** `go test ./internal/sdk/internal/hooks -run LineageSettingsStatus` and see it FAIL (undefined type).
- [ ] **Step 3: Implement the hook.** Take the structure and error handling from `compute_env_status_hook.go`. Parse `orgId` and `workspaceId` from `res.Request.URL.Path`, poll `GET {BaseURL}/orgs/{orgId}/workspaces/{workspaceId}/settings/lineage` using the request's `Authorization` header, and decode only `{status, message}`. Return when the status is `AVAILABLE`; fail on `ERRORED` using the API `message`; on timeout fail with the last status. Retry 429/5xx and network errors every `RetryInterval`. No polling on `DELETE`.
- [ ] **Step 4: Register** the hook in `initHooks` in `registration.go` with a comment in the existing style, and add both new files to `.genignore`.
- [ ] **Step 5: Run** `go test ./internal/sdk/internal/hooks` and `go build ./...`. Expected: PASS.
- [ ] **Step 6: Commit** `feat(lineage): poll AUTO provisioning status after settings update`.

---

### Task 3: Unconfigured-workspace hook

**Files:**
- Create: `internal/sdk/internal/hooks/lineage_settings_not_configured_hook.go`
- Test: `internal/sdk/internal/hooks/lineage_settings_not_configured_hook_test.go`
- Modify: `internal/sdk/internal/hooks/registration.go`, `.genignore`

**Interfaces:**
- Produces: `type LineageSettingsNotConfiguredHook struct{}` with `AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error)`. It acts only on `OperationID == "FindLineageWorkspaceSettings"` with status 200. If the body's `credentialsId` is missing or empty, it rewrites the response to 404 with JSON body `{"message":"Lineage is not configured for this workspace"}`, as `GenericResourceErrorHook` does. Otherwise it restores the original body untouched.

- [ ] **Step 1: Write failing tests:** `TestLineageSettingsNotConfiguredHook_EmptyBecomes404` (body `{"orgId":1,"wspId":2,"provisioned":false}` gives status 404), `_ConfiguredPassesThrough` (body with `credentialsId` stays 200 and the body is still readable), `_PendingWithCredentialsPassesThrough` (status `PENDING` with credentials stays 200, so `ERRORED`/`PENDING` configs stay in state), `_OtherOperationsUntouched`.
- [ ] **Step 2: Run** `go test ./internal/sdk/internal/hooks -run LineageSettingsNotConfigured`. Expected: FAIL.
- [ ] **Step 3: Implement** the hook, reading the body and restoring it on the pass-through path.
- [ ] **Step 4: Register** it after the status hook and add both files to `.genignore`.
- [ ] **Step 5: Run** `go test ./internal/sdk/internal/hooks && go build ./...`. Expected: PASS.
- [ ] **Step 6: Commit** `feat(lineage): treat an unconfigured workspace as a missing resource`.

---

### Task 4: MANUAL/AUTO field validator

**Files:**
- Create: `internal/validators/stringvalidators/lineage_manual_field_validator.go`
- Test: `internal/validators/stringvalidators/lineage_manual_field_validator_test.go`
- Modify: `overlays/workspace-lineage-settings.yaml` (add `x-speakeasy-plan-validators: LineageManualFieldValidator` on `storeName` and `topicArn`), `.genignore`

**Interfaces:**
- Produces: `func LineageManualFieldValidator() validator.String` (the name Speakeasy resolves, as with `FusionVersionValidator`). Reads the sibling `provisioning_mode` through `req.Path.ParentPath().AtName("provisioning_mode")` from `req.Config`. Null or `"AUTO"` is AUTO.
- Rules: mode `MANUAL` and the field null gives an error naming the field. Mode AUTO and the field set (not null, not unknown) gives an error saying the backend ignores it in AUTO. Unknown mode or unknown field: no error. Error summaries: "Lineage Field Required for MANUAL Provisioning" and "Lineage Field Not Allowed for AUTO Provisioning".

- [ ] **Step 1: Write the table test** in the style of `fusion_version_validator_test.go`, for the `store_name` path: `MANUAL`+null errors; `MANUAL`+set passes; `AUTO`+set errors; `AUTO`+null passes; mode null+set errors (null is AUTO); mode null+null passes; mode unknown passes; field unknown passes. Add one case for `topic_arn` to show the validator is path-agnostic.
- [ ] **Step 2: Run** `go test ./internal/validators/stringvalidators -run LineageManualField`. Expected: FAIL.
- [ ] **Step 3: Implement** the validator, following the type layout of `fusion_version_validator.go`.
- [ ] **Step 4: Add the annotation** to the overlay, add both files to `.genignore`, run `speakeasy run --skip-versioning`, and confirm `internal/provider/workspacelineagesettings_resource.go` references `custom_stringvalidators.LineageManualFieldValidator()` on both attributes.
- [ ] **Step 5: Run** `go build ./... && go test ./internal/...`. Expected: PASS.
- [ ] **Step 6: Commit** `feat(lineage): validate MANUAL/AUTO fields at plan time`.

---

### Task 5: Example, docs and plan verification

**Files:**
- Create: `examples/resources/seqera_workspace_lineage_settings/resource.tf` (AUTO and MANUAL examples; the MANUAL example wires `aws_sns_topic_subscription` to `webhook_url`)
- Modify: `.genignore` (protect the example), generated `docs/resources/workspace_lineage_settings.md`

- [ ] **Step 1: Write the example** and protect it in `.genignore`, following the pattern for other custom examples. Regenerate with `speakeasy run --skip-versioning`.
- [ ] **Step 2: Check the docs.** `git diff -- docs/` shows single-line descriptions, `webhook_url` noted as sensitive, and no `queue_arn`/`store_uri`.
- [ ] **Step 3: Run the plan checks** from `docs-internal/SPEC_UPDATE_RUNBOOK.md` section 5 with `dev_overrides` (a scratch dir). Expected outcomes: AUTO with only `credentials_id`/`region` plans; MANUAL with `store_name` and `topic_arn` plans; MANUAL without `topic_arn` fails with the "Required for MANUAL" error; AUTO with `store_name` set fails with the "Not Allowed for AUTO" error; `provisioning_mode` omitted plus `topic_arn` set fails.
- [ ] **Step 4: Run** `go build -o terraform-provider-seqera && go test ./internal/...`. Expected: PASS.
- [ ] **Step 5: Commit** `docs(lineage): add example and docs for seqera_workspace_lineage_settings`.

---

## Known deviations from the spec (to confirm with the user)

1. **Import id** is the generator's JSON object string, not `<org_id>/<workspace_id>`.
2. **"Stays in state on ERRORED/timeout"** cannot be done with a generated Create: when the hook returns an error, Terraform does not save the resource. Impact is small because `PUT` is idempotent and the next apply re-provisions an `ERRORED` AUTO config. Updates failing leave the prior state. Making it stay in state would need a hand-written resource.
3. **Not verifiable here:** behaviour against a live backend (polling timing, `webhook_url` content, the AUTO lock 400).
