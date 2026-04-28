resource "seqera_azure_batch_ce" "my_azurebatchce" {
  config = {
    auto_pool_mode                    = false
    delete_jobs_on_completion         = "always"
    delete_jobs_on_completion_enabled = true
    delete_pools_on_completion        = false
    delete_tasks_on_completion        = false
    enable_fusion                     = false
    enable_wave                       = false
    environment = [
      {
        compute = false
        head    = false
        name    = "...my_name..."
        value   = "...my_value..."
      }
    ]
    forge = {
      auto_scale = true
      container_reg_ids = [
        "..."
      ]
      dispose_on_deletion = false
      dual_pool_config    = true
      head_pool = {
        auto_scale = false
        vm_count   = 2
        vm_type    = "...my_vm_type..."
      }
      vm_count = 5
      vm_type  = "...my_vm_type..."
      worker_pool = {
        auto_scale = true
        vm_count   = 5
        vm_type    = "...my_vm_type..."
      }
    }
    head_job_cpus                     = 1
    head_job_memory_mb                = 4096
    head_pool                         = "...my_head_pool..."
    job_max_wall_clock_time           = "7d"
    managed_identity_client_id        = "...my_managed_identity_client_id..."
    managed_identity_head_resource_id = "...my_managed_identity_head_resource_id..."
    managed_identity_pool_client_id   = "...my_managed_identity_pool_client_id..."
    managed_identity_pool_resource_id = "...my_managed_identity_pool_resource_id..."
    nextflow_config                   = "...my_nextflow_config..."
    post_run_script                   = "...my_post_run_script..."
    pre_run_script                    = "...my_pre_run_script..."
    region                            = "...my_region..."
    subnet_id                         = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myRg/providers/Microsoft.Network/virtualNetworks/myVnet/subnets/mySubnet"
    terminate_jobs_on_completion      = false
    token_duration                    = "...my_token_duration..."
    work_dir                          = "az://my-container/work"
    worker_pool                       = "...my_worker_pool..."
  }
  credentials_id = "...my_credentials_id..."
  description    = "...my_description..."
  label_ids = [
    1
  ]
  name         = "...my_name..."
  platform     = "azure-batch"
  workspace_id = 1
}