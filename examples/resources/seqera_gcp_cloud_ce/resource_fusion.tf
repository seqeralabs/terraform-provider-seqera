# GCP Cloud with Fusion v2 — mounts GCS buckets as a distributed file
# system, accelerating data-heavy workloads.
#
# Fusion v2 and Wave are always on for Cloud compute environments — the
# backend enables them and they are not user-settable
resource "seqera_gcp_cloud_ce" "fusion" {
  name           = "gcp-cloud-fusion"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    project_id            = "my-gcp-project"
    region                = "us-central1"
    zone                  = "us-central1-a"
    work_dir              = "gs://my-bucket/work"
    instance_type         = "n2-standard-4"
    service_account_email = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"
    boot_disk_size_gb     = 100
  }
}
