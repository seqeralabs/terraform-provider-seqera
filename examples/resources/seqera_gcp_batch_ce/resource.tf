resource "seqera_gcp_batch_ce" "minimal" {
  name           = "gcp-batch-minimal"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    project_id      = "my-gcp-project"
    location        = "us-central1"
    work_dir        = "gs://my-bucket/work"
    machine_type    = "n1-standard-4"
    service_account = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"
  }
}
