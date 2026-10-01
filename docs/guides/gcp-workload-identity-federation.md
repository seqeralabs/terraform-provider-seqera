---
page_title: "GCP Credentials with Workload Identity Federation"
subcategory: "Examples"
description: |-
  End-to-end example of creating a Seqera Google credential that uses Workload Identity Federation to impersonate a GCP service account — no service account keys stored in Seqera.
---

# GCP Credentials with Workload Identity Federation

This guide creates a `seqera_google_credential` that authenticates to Google Cloud using [Workload Identity Federation (WIF)](https://docs.cloud.google.com/iam/docs/workload-identity-federation) instead of a long-lived service account key. The Seqera Platform acts as an OIDC identity provider: it mints a short-lived token for each operation, GCP's Security Token Service exchanges that token for a federated token, and Seqera then impersonates a GCP service account to call Batch, Storage, and other Google APIs.

~> **Note:** WIF is the recommended way to grant Seqera access to GCP, since no long-lived service account key is stored in the platform. The alternative is uploading a service account key as `data`, which is long-lived and must be rotated manually. WIF is gated behind the Identity Federation feature flag; where it is disabled, creating the credential fails at apply time.

## How the trust works

Seqera signs a token with these claims:

| Claim          | Value                                                                                                   |
| -------------- | ------------------------------------------------------------------------------------------------------- |
| `iss`          | The Seqera OIDC issuer URL, for example `https://cloud.seqera.io/api`                                   |
| `aud`          | `//iam.googleapis.com/<workload_identity_provider>` by default, or `token_audience` if you set it       |
| `sub`          | `org:<ORG_ID>:wsp:<WORKSPACE_ID>:<WORKLOAD>` for workspaces inside an org, or `usr:<USER_ID>:<WORKLOAD>` for a personal workspace |
| `principal_id` | The internal numeric ID of the acting user, on requests a user initiated                                |

The last segment of `sub` is the workload type, and **one credential presents all four**:

| Workload   | Used by                                                                                 |
| ---------- | --------------------------------------------------------------------------------------- |
| `platform` | Credential validation, compute environment describe and provisioning, job submission, run logs |
| `data`     | Data Explorer browsing and presigned download and upload URLs                           |
| `studio`   | A Studio session reading and writing its own data                                       |
| `workflow` | The bucket-reachability check before a pipeline launch                                  |

A binding that admits only one workload type breaks the others. Credential validation presents `platform`, so the credential cannot even be saved without it.

The issuer URL is the Platform's public address with `/api`, and GCP matches it byte for byte against the token's `iss`. The [`seqera_gcp_credentials_federation_setup`](../data-sources/gcp_credentials_federation_setup.md) data source returns the exact value for your installation, together with the attribute mapping and attribute condition described below, so you do not need to assemble them by hand.

### Attribute mapping and audit

GCP Cloud Audit Logs record only the mapped `google.subject`. The Platform therefore recommends this mapping, which appends the acting user to the subject:

```
google.subject = assertion.sub + (has(assertion.principal_id) ? ':usr:' + assertion.principal_id : '')
```

A user-initiated request is then logged as, for example, `org:<ORG_ID>:wsp:<WORKSPACE_ID>:data:usr:448291`. The `has()` guard keeps background tokens (cache refreshes, checkpoints, job polling), which carry no `principal_id`, valid. Data-plane activity such as Cloud Storage reads only appears in Data Access audit logs, which are off by default.

### Attribute condition

The Seqera issuer is shared by every organization and workspace on an installation. Pin the provider to your tenant with an attribute condition, otherwise a pool-wide binding also accepts tokens minted for other workspaces:

```
assertion.sub.startsWith('org:<ORG_ID>:wsp:<WORKSPACE_ID>:')
```

## Prerequisites

- Identity Federation enabled for your Seqera organization.
- A Seqera workspace you can create credentials in, and its **numeric** org ID and workspace ID (not the slug).
- A GCP project where you can create a workload identity pool and provider, a service account, and IAM bindings.
- Your project **number** (not project ID): `gcloud projects describe PROJECT_ID --format='value(projectNumber)'`.

## Step 1: Create the GCP workload identity pool and provider

Create a pool and an OIDC provider inside it that trusts Seqera as the issuer, with the attribute mapping and condition above.

```shell
# Pool — logical container for external identities.
gcloud iam workload-identity-pools create seqera-pool \
    --location=global \
    --display-name="Seqera Platform"

# OIDC provider — trusts tokens issued by the Seqera Platform.
gcloud iam workload-identity-pools providers create-oidc seqera-provider \
    --location=global \
    --workload-identity-pool=seqera-pool \
    --issuer-uri="https://cloud.seqera.io/api" \
    --attribute-mapping="google.subject=assertion.sub + (has(assertion.principal_id) ? ':usr:' + assertion.principal_id : '')" \
    --attribute-condition="assertion.sub.startsWith('org:ORG_ID:wsp:WORKSPACE_ID:')"
```

For Enterprise installs, replace `--issuer-uri` with your installation's issuer URL, for example `https://seqera.example.com/api`. Leave the allowed audiences at their default unless you set a custom `token_audience` on the credential.

~> **Note:** Pool and provider IDs cannot be changed. To rename one, create a new pool or provider instead.

## Step 2: Create the service account and let the pool impersonate it

Create the service account Seqera will impersonate, and grant `roles/iam.workloadIdentityUser` on it to the whole pool. The provider's attribute condition already restricts the pool to your tenant.

```shell
gcloud iam service-accounts create seqera-runner \
    --display-name="Seqera workflow runner"

gcloud iam service-accounts add-iam-policy-binding \
    seqera-runner@PROJECT_ID.iam.gserviceaccount.com \
    --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/seqera-pool/*"

# Presigned download URLs are signed with the IAM signBlob API, which needs
# the service account to be able to sign as itself.
gcloud iam service-accounts add-iam-policy-binding \
    seqera-runner@PROJECT_ID.iam.gserviceaccount.com \
    --role=roles/iam.serviceAccountTokenCreator \
    --member="serviceAccount:seqera-runner@PROJECT_ID.iam.gserviceaccount.com"
```

~> **Warning:** Do not bind exact subjects (`.../seqera-pool/subject/org:ORG_ID:wsp:WORKSPACE_ID:workflow`). A binding on one subject admits only that workload type, and once the attribute mapping appends the acting user, a user-initiated request's mapped subject no longer matches an exact binding at all. To scope a shared pool to individual workspaces, use an `attribute.workspace` binding instead, as shown in [Scoping a shared pool per workspace](#scoping-a-shared-pool-per-workspace).

Then grant the service account the permissions your workloads need:

- **Compute environments:** the Batch, Logging, and Compute describe permissions the compute environment needs, plus job submission. Forge-created environments also need `iam.serviceAccounts.create`, `iam.serviceAccounts.delete`, `iam.serviceAccounts.get`, `resourcemanager.projects.getIamPolicy`, `resourcemanager.projects.setIamPolicy`, and `compute.*` on the instances it creates.
- **Work directory bucket:** object read and write, for example `roles/storage.objectUser`, and `storage.objects.list`, which the check before each pipeline launch uses.
- **Data Explorer:** read on the buckets you browse, and write for uploads, for example `roles/storage.objectViewer` and `roles/storage.objectCreator`.

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

Setting `workload_identity_provider` and `service_account_email` together selects WIF mode: no `data` (service account key) is stored. The two fields must be provided together; the credential's plan validator rejects a configuration that sets only one. Set `token_audience` only when the provider is configured with a custom allowed audience.

## Step 4: Apply and use

```shell
terraform apply
```

The Platform validates the credential when it is saved by performing a live token exchange and impersonation, so a wrong issuer, mapping, condition, or binding shows up straight away: the credential is marked invalid with the reason GCP returned. See [Troubleshooting](#troubleshooting).

Reference the credential from any resource that accepts `credentials_id`, such as a Google Batch compute environment:

```terraform
resource "seqera_gcp_batch_ce" "example" {
  name           = "gcp-batch-example"
  workspace_id   = var.workspace_id
  credentials_id = seqera_google_credential.wif.credentials_id
  # ... compute env config ...
}
```

## Managing the GCP resources with the google provider

To manage the pool, provider, service account, and IAM bindings from the same Terraform configuration, pair the `seqera` provider with `hashicorp/google`. The `seqera_gcp_credentials_federation_setup` data source supplies the issuer, attribute mapping, and attribute condition for the workspace.

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

variable "gcp_project_id" {
  type = string
}

data "seqera_gcp_credentials_federation_setup" "this" {
  workspace_id = var.workspace_id
}

resource "google_iam_workload_identity_pool" "seqera" {
  workload_identity_pool_id = "seqera-pool"
  display_name              = "Seqera Platform"
}

resource "google_iam_workload_identity_pool_provider" "seqera" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.seqera.workload_identity_pool_id
  workload_identity_pool_provider_id = "seqera-provider"

  oidc {
    issuer_uri = data.seqera_gcp_credentials_federation_setup.this.oidc_issuer_url
  }

  # GCP writes only google.subject to Cloud Audit Logs, so the acting user is
  # folded into the subject rather than mapped to a custom attribute.
  attribute_mapping = {
    "google.subject" = data.seqera_gcp_credentials_federation_setup.this.google_subject_mapping
  }

  # The Seqera issuer is shared by every tenant of an installation; this pins
  # the provider to subjects minted for this workspace.
  attribute_condition = data.seqera_gcp_credentials_federation_setup.this.recommended_attribute_condition
}

resource "google_service_account" "seqera_runner" {
  account_id   = "seqera-runner"
  display_name = "Seqera workflow runner"
}

# Attach the GCP roles your workloads need here, for example:
#
# resource "google_project_iam_member" "seqera_runner_batch" {
#   project = var.gcp_project_id
#   role    = "roles/batch.jobsEditor"
#   member  = "serviceAccount:${google_service_account.seqera_runner.email}"
# }

# Every identity the pool admits may impersonate the service account; the
# provider's attribute condition limits those to this workspace.
resource "google_service_account_iam_member" "seqera_impersonate" {
  service_account_id = google_service_account.seqera_runner.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.seqera.name}/*"
}

# Presigned download URLs are signed with the IAM signBlob API.
resource "google_service_account_iam_member" "seqera_sign_blob" {
  service_account_id = google_service_account.seqera_runner.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${google_service_account.seqera_runner.email}"
}

resource "seqera_google_credential" "wif" {
  name         = "gcp-wif"
  workspace_id = var.workspace_id

  service_account_email      = google_service_account.seqera_runner.email
  workload_identity_provider = google_iam_workload_identity_pool_provider.seqera.name

  # The Platform validates the credential on save, so the bindings must be
  # live first.
  depends_on = [
    google_service_account_iam_member.seqera_impersonate,
    google_service_account_iam_member.seqera_sign_blob,
  ]
}
```

### Scoping a shared pool per workspace

When one pool serves several Seqera workspaces, pin the provider to your organization instead of a single workspace, and map the workspace ID out of the subject so each service account can be bound to one workspace:

```terraform
variable "org_id" {
  type        = number
  description = "Seqera organization ID that owns the workspaces."
}

resource "google_iam_workload_identity_pool_provider" "seqera" {
  # ... as above ...

  attribute_mapping = {
    "google.subject"      = data.seqera_gcp_credentials_federation_setup.this.google_subject_mapping
    "attribute.workspace" = "assertion.sub.extract('wsp:{workspace}:')"
  }

  attribute_condition = "assertion.sub.startsWith('org:${var.org_id}:')"
}

resource "google_service_account_iam_member" "seqera_impersonate" {
  service_account_id = google_service_account.seqera_runner.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.seqera.name}/attribute.workspace/${var.workspace_id}"
}
```

One `attribute.workspace` binding covers every workload type of that workspace. It reads the raw `sub`, so appending the acting user in `google.subject` does not affect it.

### Variant — attach to an existing workload identity pool

If the pool and provider are already managed out of band — typical when a central platform team owns WIF for the whole organization — look up the provider as a data source. You only need to own the service account and its IAM bindings, and pass the existing provider's full path into the Seqera credential.

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

The existing provider must already use the Seqera issuer URL, and its attribute condition and bindings must admit this workspace's subjects.

## Troubleshooting

| Symptom | Cause | Fix |
| ------- | ----- | --- |
| Validation fails with an issuer or discovery error | The issuer URL is wrong, missing `/api`, or not publicly reachable over HTTPS | Use `oidc_issuer_url` from the data source. GCP must be able to fetch `<issuer>/.well-known/openid-configuration` |
| Validation fails with an audience mismatch | A custom allowed audience on the provider, or a provider that was recreated under a new ID | Set `token_audience` to the provider's allowed audience, or update `workload_identity_provider` |
| The exchange succeeds but impersonation is denied | `roles/iam.workloadIdentityUser` is missing, or bound to exact subjects | Bind the whole pool or `attribute.workspace`, as in Step 2 |
| `Could not obtain a value for google.subject` | The mapping uses `principal_id` without the `has()` guard | Use the mapping from the data source |
| Presigned downloads fail to sign | The service account lacks the `serviceAccountTokenCreator` binding on itself | Add it, as in Step 2 |
| Pipeline launch refused with `WORK_DIR_INVALID` | The service account cannot list the work directory bucket | Grant `storage.objects.list` on the bucket |
| No acting user in the audit logs | The mapping does not append `principal_id`, or Data Access logs are off | Use the recommended mapping and enable Data Access audit logs |

See [Troubleshoot Workload Identity Federation](https://docs.cloud.google.com/iam/docs/troubleshooting-workload-identity-federation) for GCP's error codes.

## Notes

- The `workload_identity_provider` path uses the GCP **project number**, not the project ID.
- The subjects are derived from the workspace that owns the credential. Changing `workspace_id` forces replacement and produces new subjects, so the attribute condition and bindings must be updated in lockstep.
- Personal-workspace credentials use `usr:<USER_ID>:<WORKLOAD>` subjects, and the data source returns the matching attribute condition. Prefer org workspace credentials in production — a personal workspace binding is tied to one individual's account.
- `principal_id` is an internal numeric user ID, not an email. Background operations carry none, so some audit entries show the workspace but no user.

## Related

- Resource reference: [`seqera_google_credential`](../resources/google_credential.md)
- Data source reference: [`seqera_gcp_credentials_federation_setup`](../data-sources/gcp_credentials_federation_setup.md)
- [Configure Workload Identity Federation with other identity providers](https://docs.cloud.google.com/iam/docs/workload-identity-federation-with-other-providers) — authoritative GCP reference for the pool, provider, and IAM binding flow used above.
- [Best practices for using Workload Identity Federation](https://cloud.google.com/iam/docs/best-practices-for-using-workload-identity-federation) — audience validation, attribute conditions, and principal scoping guidance.
