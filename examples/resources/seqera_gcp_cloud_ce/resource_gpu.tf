# GCP Cloud with GPU support — for AI/ML pipelines that need accelerators
# (e.g. AlphaFold, deep-variant callers).
resource "seqera_gcp_cloud_ce" "gpu" {
  name           = "gcp-cloud-gpu"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "google-cloud"
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    project_id            = "my-gcp-project"
    region                = "us-central1"
    zone                  = "us-central1-c"
    work_dir              = "gs://my-bucket/work"
    instance_type         = "a2-highgpu-1g"
    service_account_email = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"
    gpu_enabled           = true
    boot_disk_size_gb     = 200
  }
}
