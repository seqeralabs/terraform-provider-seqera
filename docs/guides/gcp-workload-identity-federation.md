---
page_title: "GCP Credentials with Workload Identity Federation"
subcategory: "Examples"
description: |-
  End-to-end example of creating a Seqera Google credential that uses Workload Identity Federation to impersonate a GCP service account — no service account keys stored in Seqera.
---

# GCP Credentials with Workload Identity Federation

This guide creates a `seqera_google_credential` that authenticates to Google Cloud using [Workload Identity Federation (WIF)](https://docs.cloud.google.com/iam/docs/workload-identity-federation) instead of a long-lived service account key. The Seqera Platform acts as an OIDC identity provider: it mints a short-lived JWT on each workflow run, GCP's Security Token Service exchanges that JWT for a federated token, and Seqera then impersonates a GCP service account to drive Batch, Storage, and other Google APIs.

~> **Note:** WIF is the recommended way to grant Seqera access to GCP and is the most secure path, since no long-lived service account key is stored in the platform. The alternative is uploading a service account key as `data`, which is long-lived and must be rotated manually.

## How the trust works

Seqera's public API endpoint is the OIDC issuer. For Seqera Cloud this is `https://api.cloud.seqera.io`; for Enterprise installs it is the `/api` base URL of your install (whatever your browser hits when you open the platform UI, with `/api` appended).

For each workflow run, the platform signs a JWT with these claims:

| Claim        | Value                                                                                                                         |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| `iss`        | The Seqera public API endpoint, e.g. `https://api.cloud.seqera.io`                                                            |
| `aud`        | `//iam.googleapis.com/<workload_identity_provider>` by default, or `token_audience` if you set it                             |
| `sub`        | `org:<ORG_ID>:wsp:<WORKSPACE_ID>:workflow` for workspaces inside an org, or `usr:<USER_ID>:workflow` for a personal workspace |
| `iat`, `exp` | Issued-at and a one-hour expiry                                                                                               |

The `sub` claim is what GCP's IAM binding grants access to, so pinning the binding to a specific Seqera org/workspace or user is what scopes the federation. Any workflow run in a different workspace produces a different `sub` and will not match the binding.

## Prerequisites

- A Seqera Platform workspace you can create credentials in, and the **numeric** org ID and workspace ID (not the slug). Retrieve these from the workspace URL or the `seqera_workspace` data source.
- A GCP project where you can create a workload identity pool and provider, a service account, and IAM bindings.
- Your project **number** (not project ID). `gcloud projects describe PROJECT_ID --format='value(projectNumber)'`.

## Step 1: Create the GCP workload identity pool and provider

On the GCP side, create a pool, an OIDC provider inside it that trusts Seqera as the issuer, and a service account for Seqera to impersonate. The trust policy is enforced by three fields: the `issuer-uri`, the `allowed-audiences`, and the `attribute-mapping` that extracts `sub` from the Seqera JWT.

```shell
# Pool — logical container for external identities.
gcloud iam workload-identity-pools create seqera-pool \
    --location=global \
    --display-name="Seqera Platform"

# OIDC provider — trusts tokens issued by the Seqera Platform.
gcloud iam workload-identity-pools providers create-oidc seqera-provider \
    --location=global \
    --workload-identity-pool=seqera-pool \
    --issuer-uri="https://api.cloud.seqera.io" \
    --allowed-audiences="//iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/seqera-pool/providers/seqera-provider" \
    --attribute-mapping="google.subject=assertion.sub"

# Service account Seqera will impersonate. Grant it the GCP roles your
# pipelines need (e.g. roles/batch.jobsEditor, roles/storage.objectAdmin).
gcloud iam service-accounts create seqera-runner \
    --display-name="Seqera workflow runner"
```

Substitute `PROJECT_NUMBER` in `--allowed-audiences` for your project number. The audience must exactly match the value Seqera puts in the JWT `aud` claim, which defaults to `//iam.googleapis.com/<workload_identity_provider>`.

For Enterprise installs, replace `--issuer-uri` with the `/api` endpoint of your install, for example `https://seqera.example.com/api`.

## Step 2: Grant the Seqera subject permission to impersonate the service account

Bind the specific Seqera subject that will run workflows to `roles/iam.workloadIdentityUser` on the service account.

For a workspace inside an organisation:

```shell
gcloud iam service-accounts add-iam-policy-binding \
    seqera-runner@PROJECT_ID.iam.gserviceaccount.com \
    --role=roles/iam.workloadIdentityUser \
    --member="principal://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/seqera-pool/subject/org:ORG_ID:wsp:WORKSPACE_ID:workflow"
```

For a personal workspace (no org):

```shell
gcloud iam service-accounts add-iam-policy-binding \
    seqera-runner@PROJECT_ID.iam.gserviceaccount.com \
    --role=roles/iam.workloadIdentityUser \
    --member="principal://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/seqera-pool/subject/usr:USER_ID:workflow"
```

A `principalSet://` binding on `attribute.*` allows any workspace in the pool to use the service account. This is typically too broad; prefer per-workspace bindings and add more as more workspaces need access.

## Step 3: Define the Seqera credential

```terraform
terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
  }
}

variable "workspace_id" {
  type        = number
  description = "Seqera workspace ID that will own the credential."
}

variable "gcp_project_number" {
  type        = string
  description = "Numeric GCP project number that hosts the workload identity pool."
}

variable "service_account_email" {
  type        = string
  description = "GCP service account Seqera will impersonate."
}

resource "seqera_google_credential" "wif" {
  name         = "gcp-wif"
  workspace_id = var.workspace_id

  service_account_email      = var.service_account_email
  workload_identity_provider = "projects/${var.gcp_project_number}/locations/global/workloadIdentityPools/seqera-pool/providers/seqera-provider"
}
```

Setting `workload_identity_provider` and `service_account_email` together selects WIF mode — the platform mints OIDC tokens and exchanges them via GCP STS, and no `data` (service account key) is stored. The two fields must be provided together; the credential's plan validator rejects a configuration that sets only one. If you need to override the audience embedded in the JWT — for example when fronting multiple pools with the same Seqera workspace — set `token_audience`.

## Step 4: Apply and use

```shell
terraform apply
```

Once the credential is created, reference it from any resource that accepts `credentials_id`, such as a Google Batch compute environment:

```terraform
resource "seqera_google_batch_ce" "example" {
  name           = "gcp-batch-example"
  workspace_id   = var.workspace_id
  credentials_id = seqera_google_credential.wif.credentials_id
  # ... compute env config ...
}
```

The first workflow run will surface any trust-policy mismatch — GCP returns a clear error message that names the failing claim (`iss`, `aud`, or `sub`). See [Troubleshoot Workload Identity Federation](https://docs.cloud.google.com/iam/docs/troubleshooting-workload-identity-federation) for a list of error codes.

## Managing the GCP resources with the google provider

The gcloud flow above requires operator access to the GCP project. To manage the pool, provider, service account, and IAM binding from the same Terraform configuration, pair the `seqera` provider with `hashicorp/google`.

The order of operations matches the manual flow:

1. Terraform creates the workload identity pool and OIDC provider, trusting the Seqera issuer.
2. Terraform creates the service account Seqera will impersonate, and attaches any project-level roles the service account needs for your workloads.
3. Terraform creates the IAM binding on the service account, granting `roles/iam.workloadIdentityUser` to the exact Seqera subject that will federate in.
4. Terraform creates `seqera_google_credential` with the provider path and service account email. Downstream resources that trigger WIF token exchange (compute environments and similar) set `depends_on = [google_service_account_iam_member.seqera_impersonate]` so the binding is live before Seqera uses the credential.

```terraform
terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

variable "workspace_id" {
  type = number
}

variable "org_id" {
  type        = number
  description = "Seqera organization ID that owns the workspace."
}

variable "gcp_project_id" {
  type = string
}

variable "seqera_issuer" {
  type        = string
  default     = "https://api.cloud.seqera.io"
  description = "Seqera Platform OIDC issuer. Use the /api base URL for Enterprise installs."
}

data "google_project" "current" {
  project_id = var.gcp_project_id
}

locals {
  pool_id     = "seqera-pool"
  provider_id = "seqera-provider"

  provider_path = "projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${local.pool_id}/providers/${local.provider_id}"

  seqera_subject = "org:${var.org_id}:wsp:${var.workspace_id}:workflow"
}

resource "google_iam_workload_identity_pool" "seqera" {
  workload_identity_pool_id = local.pool_id
  display_name              = "Seqera Platform"
}

resource "google_iam_workload_identity_pool_provider" "seqera" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.seqera.workload_identity_pool_id
  workload_identity_pool_provider_id = local.provider_id

  oidc {
    issuer_uri        = var.seqera_issuer
    allowed_audiences = ["//iam.googleapis.com/${local.provider_path}"]
  }

  attribute_mapping = {
    "google.subject" = "assertion.sub"
  }
}

resource "google_service_account" "seqera_runner" {
  account_id   = "seqera-runner"
  display_name = "Seqera workflow runner"
}

# Attach the GCP roles your workloads need here — see Google Batch and
# Cloud Storage permission references for the exact role set.
#
# resource "google_project_iam_member" "seqera_runner_batch" {
#   project = var.gcp_project_id
#   role    = "roles/batch.jobsEditor"
#   member  = "serviceAccount:${google_service_account.seqera_runner.email}"
# }

resource "google_service_account_iam_member" "seqera_impersonate" {
  service_account_id = google_service_account.seqera_runner.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${local.pool_id}/subject/${local.seqera_subject}"
}

resource "seqera_google_credential" "wif" {
  name         = "gcp-wif"
  workspace_id = var.workspace_id

  service_account_email      = google_service_account.seqera_runner.email
  workload_identity_provider = local.provider_path

  depends_on = [
    google_iam_workload_identity_pool_provider.seqera,
    google_service_account_iam_member.seqera_impersonate,
  ]
}
```

### Variant — attach to an existing workload identity pool

If the pool and provider are already managed out of band — typical when a central platform team owns WIF for the whole organization — skip creating them and look them up as data sources. You only need to own the service account and its IAM binding, and pass the existing provider's full path into the Seqera credential.

```terraform
variable "existing_pool_id" {
  type = string
}

variable "existing_provider_id" {
  type = string
}

data "google_iam_workload_identity_pool_provider" "existing" {
  workload_identity_pool_id          = var.existing_pool_id
  workload_identity_pool_provider_id = var.existing_provider_id
}

resource "seqera_google_credential" "wif" {
  name         = "gcp-wif"
  workspace_id = var.workspace_id

  service_account_email      = google_service_account.seqera_runner.email
  workload_identity_provider = data.google_iam_workload_identity_pool_provider.existing.name
}
```

The existing provider's `issuer-uri` must already be set to the Seqera issuer (`https://api.cloud.seqera.io` for Cloud, your `/api` endpoint for Enterprise). If the central team owns a single shared provider for multiple external issuers, they will need to confirm Seqera is one of them.

## Notes

- The `workload_identity_provider` path uses the GCP **project number**, not the project ID. Passing the project ID silently produces tokens with an `aud` that GCP cannot match.
- The `sub` claim is derived from the workspace that owns the credential. Moving the credential between workspaces — or changing `workspace_id` — forces replacement and produces a new subject, so the GCP IAM binding must be updated in lockstep.
- Personal-workspace credentials use the `usr:<USER_ID>:workflow` subject. Prefer org/workspace credentials in production — a personal workspace binding is tied to one individual's account.
- `token_audience` is an advanced override. The default `//iam.googleapis.com/<workload_identity_provider>` is the value GCP's `allowed-audiences` check expects; set a custom audience only when fronting multiple pools with the same credential.
- Credentials are tied to the workspace they were created in. Changing `workspace_id` forces replacement.

## Related

- Resource reference: [`seqera_google_credential`](../resources/google_credential.md)
- [Configure Workload Identity Federation with other identity providers](https://docs.cloud.google.com/iam/docs/workload-identity-federation-with-other-providers) — authoritative GCP reference for the pool, provider, and IAM binding flow used above.
- [Best practices for using Workload Identity Federation](https://cloud.google.com/iam/docs/best-practices-for-using-workload-identity-federation) — audience validation, attribute conditions, and principal scoping guidance.
