# GCP Batch with Fusion v2 and Wave — accelerates GCS-heavy workloads
# by mounting buckets as a distributed file system. Fusion v2 requires Wave.
resource "seqera_gcp_batch_ce" "fusion" {
  name           = "gcp-batch-fusion"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    location        = "us-central1"
    work_dir        = "gs://my-bucket/work"
    machine_type    = "n2-standard-4"
    service_account = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"
    enable_wave     = true
    enable_fusion   = true
    spot            = true
  }
}
