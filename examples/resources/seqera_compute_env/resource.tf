resource "seqera_compute_env" "my_computeenv" {
  compute_env = {
    config = {
      moab_platform = {
        compute_queue = "...my_compute_queue..."
        environment = [
          {
            compute = false
            head    = false
            name    = "...my_name..."
            value   = "...my_value..."
          }
        ]
        head_job_options           = "...my_head_job_options..."
        head_queue                 = "...my_head_queue..."
        host_name                  = "...my_host_name..."
        launch_dir                 = "...my_launch_dir..."
        max_queue_size             = 8
        nextflow_config            = "...my_nextflow_config..."
        port                       = 3
        post_run_script            = "...my_post_run_script..."
        pre_run_script             = "...my_pre_run_script..."
        propagate_head_job_options = false
        user_name                  = "...my_user_name..."
        work_dir                   = "...my_work_dir..."
      }
    }
    credentials_id = "...my_credentials_id..."
    description    = "...my_description..."
    message        = "...my_message..."
    name           = "...my_name..."
    platform       = "google-lifesciences"
  }
  force = true
  label_ids = [
    6
  ]
  workspace_id = 1
}