# GCP Cloud with Fusion v2 and Wave — mounts GCS buckets as a distributed
# file system, accelerating data-heavy workloads. Fusion v2 requires Wave.
resource "seqera_gcp_cloud_ce" "fusion" {
  name           = "gcp-cloud-fusion"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "google-cloud"
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    project_id            = "my-gcp-project"
    region                = "us-central1"
    zone                  = "us-central1-a"
    work_dir              = "gs://my-bucket/work"
    instance_type         = "n2-standard-4"
    service_account_email = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"
    enable_wave           = true
    enable_fusion         = true
    boot_disk_size_gb     = 100
  }
}
