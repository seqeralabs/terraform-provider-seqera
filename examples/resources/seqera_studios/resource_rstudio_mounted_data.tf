# Fetch all data links in the workspace
data "seqera_data_links" "workspace_data" {
  workspace_id = seqera_workspace.my_workspace.id
}

# Create a lookup map indexed by data link name
locals {
  data_links = {
    for dl in data.seqera_data_links.workspace_data.data_links : dl.name => dl
  }
}

resource "seqera_studios" "rstudio_with_data" {
  auto_start     = false
  compute_env_id = "htaAEef9YYm5DqQrAyeDy"
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
  workspace_id         = seqera_workspace.my_workspace.id
}

# Alternative: Mount only AWS data links in us-east-1
resource "seqera_studios" "rstudio_regional_data" {
  auto_start     = false
  compute_env_id = "htaAEef9YYm5DqQrAyeDy"
  configuration = {
    cpu            = 2
    memory         = 8192
    lifespan_hours = 8
    # Filter and mount only AWS data links in us-east-1
    mount_data = [
      for dl in data.seqera_data_links.workspace_data.data_links :
      dl.id if dl.provider == "aws" && dl.region == "us-east-1"
    ]
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "cr.seqera.io/public/data-studio-ride:2025.04.1-snapshot"
  description          = "RStudio with AWS us-east-1 data only"
  is_private           = true
  name                 = "rstudio-regional-data"
  workspace_id         = seqera_workspace.my_workspace.id
}
