# The values GCP needs before a workload-identity credential can be created.
# Resolved for a workspace context; omit workspace_id for the user context.
data "seqera_gcp_credentials_federation_setup" "this" {
  workspace_id = seqera_workspace.my_workspace.id
}

resource "google_iam_workload_identity_pool" "seqera" {
  workload_identity_pool_id = "seqera-platform"
}

resource "google_iam_workload_identity_pool_provider" "seqera" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.seqera.workload_identity_pool_id
  workload_identity_pool_provider_id = "seqera-oidc"

  oidc {
    issuer_uri = data.seqera_gcp_credentials_federation_setup.this.oidc_issuer_url
  }

  # GCP writes only google.subject to Cloud Audit Logs, so the acting Platform
  # user is folded into the subject rather than mapped to a custom attribute.
  attribute_mapping = {
    "google.subject" = data.seqera_gcp_credentials_federation_setup.this.google_subject_mapping
  }

  # Seqera is a shared issuer across every tenant of an installation. Without
  # this condition the pool would accept a subject minted for any other
  # organization or workspace.
  attribute_condition = data.seqera_gcp_credentials_federation_setup.this.recommended_attribute_condition
}

# Feed the provider back into the credential that federates through it.
resource "seqera_google_credential" "workload_identity" {
  name                        = "gcp-workload-identity"
  workspace_id                = seqera_workspace.my_workspace.id
  workload_identity_provider  = google_iam_workload_identity_pool_provider.seqera.name
  service_account_email       = google_service_account.seqera.email
}

# Every value as returned by the API, including any label not mapped above.
output "all_setup_values" {
  value = data.seqera_gcp_credentials_federation_setup.this.setup_values
}
