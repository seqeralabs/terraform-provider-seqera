# Basic Google Batch compute environment example
resource "seqera_google_batch_ce" "basic" {
  name           = "gcp-batch-basic"
  workspace_id   = 123
  credentials_id = seqera_google_credential.main.id
  description    = "Basic Google Batch compute environment"

  config = {
    location = "us-central1"
    work_dir = "gs://my-bucket/work"
  }
}

# Google Batch with Spot instances and Fusion
resource "seqera_google_batch_ce" "spot_fusion" {
  name           = "gcp-batch-spot"
  workspace_id   = 123
  credentials_id = seqera_google_credential.main.id
  description    = "Google Batch with Spot and Fusion"

  config = {
    location            = "us-central1"
    work_dir            = "gs://my-bucket/work"
    spot                = true
    boot_disk_size_gb   = 200
    use_private_address = false
    head_job_cpus       = 2
    head_job_memory_mb  = 4096
    enable_wave         = true
    enable_fusion       = true

    environment = []
  }
}

# Google Batch with custom configuration
resource "seqera_google_batch_ce" "advanced" {
  name           = "gcp-batch-advanced"
  workspace_id   = 123
  credentials_id = seqera_google_credential.main.id
  description    = "Advanced Google Batch configuration"

  config = {
    location          = "us-central1"
    work_dir          = "gs://my-bucket/work"
    boot_disk_size_gb = 100

    # Network configuration
    network             = "projects/my-project/global/networks/default"
    subnetwork          = "projects/my-project/regions/us-central1/subnetworks/default"
    use_private_address = false

    # Service account
    service_account = "nextflow-sa@my-project.iam.gserviceaccount.com"

    # Resource labels
    labels = {
      environment = "production"
      team        = "data-science"
    }

    # Head job configuration
    head_job_cpus      = 4
    head_job_memory_mb = 8192

    # Wave and Fusion
    enable_wave   = true
    enable_fusion = true

    # Scripts
    pre_run_script = <<-EOT
      #!/bin/bash
      echo "Setting up environment..."
    EOT

    post_run_script = <<-EOT
      #!/bin/bash
      echo "Cleaning up..."
    EOT

    # Environment variables
    environment = [
      {
        name    = "MY_VAR"
        value   = "my_value"
        head    = true
        compute = true
      }
    ]
  }
}
