resource "seqera_managed_compute_ce" "my_managedcomputece" {
  data_retention_policy = true
  environment = [
    {
      compute = true
      head    = false
      name    = "...my_name..."
      value   = "...my_value..."
    }
  ]
  instance_size = "SMALL"
  label_ids = [
    6
  ]
  name            = "...my_name..."
  nextflow_config = "...my_nextflow_config..."
  post_run_script = "...my_post_run_script..."
  pre_run_script  = "...my_pre_run_script..."
  region          = "us-east-1"
  resource_label_ids = [
    1
  ]
  work_dir     = "work"
  workspace_id = 10
}