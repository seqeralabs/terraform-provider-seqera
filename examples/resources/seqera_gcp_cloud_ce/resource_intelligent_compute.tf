# Seqera Intelligent Compute (Preview) distributes tasks across multiple
# Compute Engine VMs with optimised scheduling.
resource "seqera_gcp_cloud_ce" "intelligent" {
  name           = "gcp-cloud-intelligent"
  workspace_id   = data.seqera_workspace.main.id
  credentials_id = seqera_google_credential.main.credentials_id

  config = {
    project_id            = "my-gcp-project"
    region                = "us-central1"
    zone                  = "us-central1-a"
    work_dir              = "gs://my-bucket/work"
    instance_type         = "n2-standard-4"
    service_account_email = "seqera-runner@my-gcp-project.iam.gserviceaccount.com"

    intelligent_compute_enabled = true
    intelligent_compute_config = {
      provisioning_model = "spotFirst" # spot | spotFirst | ondemand

      # Restrict the machine types the scheduler may use. Omit to let it pick
      # the most cost-optimal types per task. Each entry is a machine family,
      # an exact machine type, or a glob with `*`.
      machine_types = [
        "n2",            # every n2 machine type
        "c3-standard-8", # exactly this machine type
        "e2-highmem-*",  # every e2 high-memory machine type
      ]
    }
  }
}
