# Fetch all data links in the workspace
data "seqera_data_links" "workspace_data" {
  workspace_id = seqera_workspace.main.id
}

# Create a lookup map indexed by data link name
locals {
  data_links = {
    for dl in data.seqera_data_links.workspace_data.data_links : dl.name => dl
  }
}

resource "seqera_studios" "rstudio_with_data" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
  configuration = {
    cpu            = 2
    memory         = 8192
    lifespan_hours = 8
    # Mount data links by referencing them by name from the datasource
    # This allows you to dynamically reference S3/Azure/GCS buckets configured in your workspace
    mount_data = [
      local.data_links["my-s3-bucket"].id,
      local.data_links["my-analysis-data"].id,
    ]
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "cr.seqera.io/public/data-studio-ride:2025.04.1-snapshot"
  description          = "RStudio with mounted S3 data"
  is_private           = true
  name                 = "rstudio-with-data"
  workspace_id         = seqera_workspace.main.id
}
