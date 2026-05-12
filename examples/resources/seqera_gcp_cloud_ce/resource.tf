# Minimal GCP Cloud compute environment.
# Nextflow runs directly on Compute Engine VMs managed by Seqera.
resource "seqera_gcp_cloud_ce" "minimal" {
  name           = "gcp-cloud-minimal"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "google-cloud"
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    project_id            = "my-gcp-project"
    region                = "us-central1"
    zone                  = "us-central1-a"
    work_dir              = "gs://my-bucket/work"
    instance_type         = "n1-standard-4"
    service_account_email = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"
  }
}
