resource "seqera_gcp_cloud_ce" "my_gcpcloudce" {
  config = {
    arm64_enabled     = false
    boot_disk_size_gb = 50
    enable_fusion     = true
    enable_wave       = false
    environment = [
      {
        compute = false
        head    = false
        name    = "...my_name..."
        value   = "...my_value..."
      }
    ]
    gpu_enabled           = false
    image_id              = "...my_image_id..."
    instance_type         = "n1-standard-4"
    nextflow_config       = "...my_nextflow_config..."
    post_run_script       = "...my_post_run_script..."
    pre_run_script        = "...my_pre_run_script..."
    project_id            = "my-gcp-project"
    region                = "us-central1"
    service_account_email = "my-sa@my-project.iam.gserviceaccount.com"
    work_dir              = "gs://my-nextflow-bucket/work"
    zone                  = "us-central1-a"
  }
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  label_ids = [
    3
  ]
  name         = "...my_name..."
  platform     = "google-cloud"
  workspace_id = 4
}