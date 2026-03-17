# Fetch all data links in the workspace
data "seqera_data_links" "workspace_data" {
  workspace_id = seqera_workspace.main.id
}

# Mount only AWS data links in a specific region
resource "seqera_studios" "rstudio_regional_data" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
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
  workspace_id         = seqera_workspace.main.id
}
