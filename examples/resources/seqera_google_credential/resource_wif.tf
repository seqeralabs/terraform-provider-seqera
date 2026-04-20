resource "seqera_google_credential" "wif" {
  name         = "gcp-wif"
  workspace_id = seqera_workspace.main.id

  workload_identity_provider = "projects/123456789012/locations/global/workloadIdentityPools/seqera-pool/providers/seqera-provider"
  service_account_email      = "seqera-runner@my-project.iam.gserviceaccount.com"
}
