# Workspace lineage settings resource — design

Date: 2026-10-20. Branch: `feat/lineage-workspace-settings-resource` (on top of
`feat/update-api-1.244.0`, which vendored API 1.244.0 and generated SDK clients
for the lineage endpoints but no resource).

## Intent

Let users configure a workspace's lineage settings from Terraform through a
dedicated resource, `seqera_workspace_lineage_settings`. Today the only way is
calling `PUT /orgs/{orgId}/workspaces/{workspaceId}/settings/lineage` by hand.

A standalone resource (not extra attributes on `seqera_workspace`) because the
settings have their own lifecycle (delete means "disable lineage"), need an AWS
credential, region and bucket/topic that do not belong on a workspace, and are a
singleton per workspace.

## Goals

- One resource, `seqera_workspace_lineage_settings`, covering both provisioning
  modes: `AUTO` (Platform creates bucket, SNS topic and subscription) and
  `MANUAL` (customer pre-creates bucket and topic).
- Terraform create/update/destroy map to the API's `PUT`/`PUT`/`DELETE`; read
  maps to `GET`. Import supported using the id `<org_id>/<workspace_id>`.
- Plan-time validation of the rules the backend enforces (below), so users see
  errors before apply.
- Expose `webhook_url` so users can wire an `aws_sns_topic_subscription` in
  `MANUAL` mode.
- Standard repo workflow: spec annotations in `schemas/seqera-final.yaml`,
  overlay in `overlays/`, hand-written pieces in `.genignore`, a custom example
  in `examples/resources/seqera_workspace_lineage_settings/resource.tf`.

## Non-goals

- A data source (can follow later; not requested).
- Managing the AWS infrastructure in `MANUAL` mode (bucket, topic, subscription).
- Destroying AWS infrastructure on delete: the API leaves it intact and so do we.
- Exposing `queue_arn` (deprecated SQS transport; the API rejects it on input)
  or `store_uri` (the backend always returns null).
- Changes to `seqera_workspace`.

## Anti-goals

- A resource that reports success while the workspace indexes nothing (for
  example `AUTO` provisioning that ended `ERRORED`).
- Perpetual diffs: computed values must not churn the plan.
- Forcing replacement for changes the API can apply in place.

## Backend behaviour the design relies on

Verified in the platform repo (`WorkspaceController`, `LineageSettingsTx`,
`LineageWorkspaceSettingsService`).

- **Upsert.** Only `PUT` (204, no body), `GET` and `DELETE` (204) exist. There
  is no create call.
- **Required input.** `credentials_id` and `region` always. In `MANUAL`,
  `store_name` and `topic_arn` too. `provisioning_mode` defaults to `AUTO` when
  omitted. `queue_arn` is rejected with 400.
- **Unconfigured workspace.** `GET` returns 200 with empty fields (null
  `credentialsId`, `provisioned=false`), not 404. Read must treat this as the
  resource being gone and remove it from state. A missing workspace is 404.
- **Delete is idempotent.** With no config it is a no-op returning 204.
- **AUTO is asynchronous.** After `PUT` the status is `PENDING`; Platform then
  provisions and moves to `AVAILABLE` or `ERRORED` (with `message`). The bucket
  name is always `seqera-lineage-{workspaceId}`; a supplied `store_name` is
  ignored.
- **AUTO lock.** Once an AUTO config is provisioned, changing `credentials_id`
  or `region` while staying AUTO is rejected with 400 ("locked for
  AUTO-provisioned workspaces"). Switching AUTO -> MANUAL is allowed and
  decommissions the Platform-managed AUTO infrastructure.
- **MANUAL is synchronous.** Status goes straight to `AVAILABLE`, but
  `event_delivery` stays `PENDING_CONFIRMATION` until the customer's SNS
  subscription to the Platform webhook is confirmed.
- **Webhook URL.** Generated from a token created on first save, so it is only
  known after the first apply.
- **Provider.** Derived from the credential (`aws` is what is supported today).

## Decided

- Resource name `seqera_workspace_lineage_settings`; id is
  `<org_id>/<workspace_id>`. Identity inputs: `org_id`, `workspace_id`
  (replacement on change).
- Inputs: `credentials_id`, `region`, `provisioning_mode` (`AUTO`/`MANUAL`,
  optional), `lineage_default` (optional boolean), `store_name` and `topic_arn`
  (required in `MANUAL`).
- Read-only: `status`, `message`, `event_delivery`, `webhook_url`, `provider`,
  `provisioned`. In `AUTO` mode `store_name` and `topic_arn` are computed from
  the API response.
- Create/update waits for AUTO provisioning to leave `PENDING`, failing with the
  API's `message` on `ERRORED`. It does **not** wait on `event_delivery`.
- Changing `credentials_id`/`region` in place is allowed at plan time; the API's
  400 is surfaced when the AUTO lock applies (no `RequiresReplace`, which would
  delete the config and re-provision).
- Cross-field plan validation: `store_name` and `topic_arn` required when mode is
  `MANUAL`; `topic_arn` not accepted in `AUTO` (the backend ignores it).
- Destroy calls `DELETE` and leaves AWS infrastructure in place; document this.

## Calls made by the author (not yet confirmed by the user)

- **Read-back.** `GET` returns `provisioningMode`, `lineageDefault`, `storeName`
  and `topicArn` (verified in `WorkspaceController.getLineageSettings`), so read
  populates every input and import can fill them all.
- **Mode switch and validators.** `store_name` and `topic_arn` are
  Optional+Computed. Validators look at the *configured* value only, never
  state, so a value computed in state after an AUTO apply cannot trip them.
  `store_name` and `topic_arn` set in config while the mode is `AUTO` are a plan
  error (the backend would ignore them and cause an inconsistent-result error).
  Switching AUTO -> MANUAL therefore requires the user to set both explicitly;
  the old computed values must not be carried over (no `UseStateForUnknown` on
  these two).
- **`provisioning_mode`.** Optional+Computed with default `AUTO`, so it is never
  null in plan or state.
- **"Gone" detection.** Read removes the resource from state only when the API
  returns no `credentialsId` (unconfigured). `PENDING` and `ERRORED` configs stay
  in state.
- **Provisioning wait.** Create/update polls until the status leaves `PENDING`.
  On `ERRORED` or timeout the apply fails with the API `message` (or a timeout
  error), but the resource stays in state so the next plan/apply re-reads it.
  Fixed 10-minute timeout, no `timeouts` block in the first version.
- **`webhook_url`.** Marked Sensitive, since it embeds a token.
- **Generated vs hand-written.** The `PUT` returns 204 with no body, so create
  and update must follow with a `GET`. If Speakeasy cannot do that cleanly, a
  hand-written resource protected by `.genignore` is acceptable.
- **Ids.** `org_id` and `workspace_id` are int64, as elsewhere in the repo. The
  builder picks the import helper for the composite id.

## Left to the builder

- How to express an upsert-backed entity in Speakeasy (one `PUT` bound to both
  `#create` and `#update`, versus a hand-written resource file protected by
  `.genignore`). Prefer the generated route if it works; see
  `overlays/primary-compute-env.yaml` and `docs-internal/OVERLAY_GUIDE.md`.
- Where polling lives (an SDK hook under `internal/sdk/internal/hooks`, like
  `compute_env_status_hook.go`, or in the resource) and its interval/timeout.
- Exact validator placement and names, and whether `provider`/`provisioned` are
  worth keeping in the schema.
- Tests: unit tests for the validators and the polling logic, plus the local
  `dev_overrides` plan check from `docs-internal/SPEC_UPDATE_RUNBOOK.md`.

## Constraints

- Generated code must not be hand-edited; hand-written files go in `.genignore`.
- Keep attribute descriptions single-line (docs list rendering).
- Follow `docs-internal/OVERLAY_GUIDE.md` and, if a schema version bump becomes
  necessary, `docs-internal/STATE_UPGRADER_GUIDE.md` (not expected: new resource).
- Cannot be applied against a real backend in this environment; verification is
  build, unit tests and `terraform plan` only.
