resource "seqera_compute_env" "my_computeenv" {
  compute_env = {
    config = {
      azure_batch = {
        auto_pool_mode             = false
        delete_jobs_on_completion  = "on_success"
        delete_pools_on_completion = false
        environment = [
          {
            compute = false
            head    = true
            name    = "...my_name..."
            value   = "...my_value..."
          }
        ]
        forge = {
          auto_scale = false
          container_reg_ids = [
            "..."
          ]
          dispose_on_deletion = true
          vm_count            = 2
          vm_type             = "...my_vm_type..."
        }
        fusion2_enabled            = false
        head_pool                  = "...my_head_pool..."
        managed_identity_client_id = "...my_managed_identity_client_id..."
        nextflow_config            = "...my_nextflow_config..."
        post_run_script            = "...my_post_run_script..."
        pre_run_script             = "...my_pre_run_script..."
        region                     = "...my_region..."
        token_duration             = "...my_token_duration..."
        wave_enabled               = true
        work_dir                   = "...my_work_dir..."
      }
    }
    credentials_id = "...my_credentials_id..."
    description    = "...my_description..."
    message        = "...my_message..."
    name           = "...my_name..."
    platform       = "google-lifesciences"
    primary        = false
  }
  label_ids = [
    6
  ]
  workspace_id = 1
}