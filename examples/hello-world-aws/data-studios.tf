resource "seqera_data_studios" "my_datastudios" {
  compute_env_id        = resource.seqera_compute_env.aws_batch_compute_env.compute_env_id
  description           = "Data studio"
  name                  = "Terraform-Data-Studio"
  configuration = {
    conda_environment = ""
    # cpu               = 6
    # gpu               = 8
    # lifespan_hours    = 5
    # memory            = 9
    # mount_data = [
    #   "..."
    # ]
  }
  #spot                  = true
  workspace_id          = resource.seqera_workspace.my_workspace.id
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
}

