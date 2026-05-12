# GCP Batch on a private VPC with no external IP — for VPC-SC perimeters
# or environments that require traffic to stay inside the corporate network.
resource "seqera_gcp_batch_ce" "private_network" {
  name           = "gcp-batch-private"
  workspace_id   = data.seqera_workspace.main.id
  platform       = "google-batch"
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    project_id          = "my-gcp-project"
    location            = "us-central1"
    work_dir            = "gs://my-bucket/work"
    machine_type        = "n1-standard-4"
    service_account     = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"
    network             = "projects/my-gcp-project/global/networks/seqera-vpc"
    subnetwork          = "projects/my-gcp-project/regions/us-central1/subnetworks/seqera-private"
    use_private_address = true
    network_tags        = ["seqera", "no-external-ip"]
  }
}
