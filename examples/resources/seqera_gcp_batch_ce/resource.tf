resource "seqera_gcp_batch_ce" "my_gcpbatchce" {
  config = {
    boot_disk_image                = "...my_boot_disk_image..."
    boot_disk_size_gb              = 50
    compute_jobs_instance_template = "...my_compute_jobs_instance_template..."
    compute_jobs_machine_type = [
      "..."
    ]
    copy_image    = "...my_copy_image..."
    cpu_platform  = "...my_cpu_platform..."
    debug_mode    = 0
    enable_fusion = true
    enable_wave   = true
    environment = [
      {
        compute = false
        head    = false
        name    = "...my_name..."
        value   = "...my_value..."
      }
    ]
    fusion_snapshots           = false
    head_job_cpus              = 4
    head_job_instance_template = "...my_head_job_instance_template..."
    head_job_memory_mb         = 8192
    labels = {
      key = "value"
    }
    location     = "us-central1"
    machine_type = "n1-standard-4"
    network      = "default"
    network_tags = [
      "..."
    ]
    nextflow_config     = "...my_nextflow_config..."
    nfs_mount           = "/mnt/nfs"
    nfs_target          = "...my_nfs_target..."
    post_run_script     = "...my_post_run_script..."
    pre_run_script      = "...my_pre_run_script..."
    project_id          = "my-gcp-project"
    service_account     = "my-sa@my-project.iam.gserviceaccount.com"
    spot                = false
    ssh_daemon          = false
    ssh_image           = "...my_ssh_image..."
    subnetwork          = "...my_subnetwork..."
    use_private_address = false
    work_dir            = "gs://my-nextflow-bucket/work"
  }
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  label_ids = [
    5
  ]
  name         = "...my_name..."
  platform     = "google-batch"
  workspace_id = 1
}