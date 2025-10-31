# Example 1: Minimal Jupyter Studio
resource "seqera_studios" "basic_jupyter" {
  name                 = "my-jupyter-studio"
  compute_env_id       = "compute-env-id"
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  workspace_id         = seqera_workspace.my_workspace.id
  # Configuration is required - gpu defaults to 0
  configuration = {}
}

# Example 2: Jupyter Studio with Conda Environment using heredoc
resource "seqera_studios" "jupyter_with_conda_heredoc" {
  auto_start     = false
  compute_env_id = "compute-env-id"
  configuration = {
    # Use heredoc for simple YAML - just copy/paste your conda environment
    conda_environment = <<-EOT
      channels:
        - conda-forge
        - bioconda
      dependencies:
        - numpy>1.7,<2.3
        - scipy
        - tqdm=4.*
        - pip:
          - matplotlib==3.10.*
          - seaborn>=0.13
    EOT
    cpu            = 2
    memory         = 4096
    lifespan_hours = 8
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  description          = "Jupyter studio with conda packages defined using heredoc"
  is_private           = true
  name                 = "jupyter-with-conda-heredoc"
  spot                 = true
  workspace_id         = seqera_workspace.my_workspace.id
}

# Example 3: Jupyter Studio with Conda Environment and Labels using yamlencode
resource "seqera_labels" "environment_prod" {
  workspace_id = seqera_workspace.my_workspace.id
  name         = "environment"
  value        = "production"
  resource     = true
}

resource "seqera_labels" "team_datascience" {
  workspace_id = seqera_workspace.my_workspace.id
  name         = "team"
  value        = "data-science"
  resource     = true
}

resource "seqera_studios" "jupyter_with_conda_labels" {
  auto_start     = false
  compute_env_id = "compute-env-id"
  configuration = {
    # Use yamlencode() for dynamic generation or when using Terraform variables
    conda_environment = yamlencode({
      channels = [
        "conda-forge",
        "bioconda"
      ]
      dependencies = [
        "numpy>1.7,<2.3",
        "scipy",
        "tqdm=4.*",
        {
          pip = [
            "matplotlib==3.10.*",
            "seaborn>=0.13"
          ]
        }
      ]
    })
    cpu            = 2
    memory         = 4096
    lifespan_hours = 8
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url  = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  description           = "Jupyter studio for data analysis and visualization"
  is_private            = true
  # Reference label IDs from seqera_labels resources
  label_ids = [
    seqera_labels.environment_prod.id,
    seqera_labels.team_datascience.id
  ]
  name         = "jupyter-with-conda-labels"
  spot         = true
  workspace_id = seqera_workspace.my_workspace.id
}


# Example 4: RStudio with Mounted Data using Data Links Datasource
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

# Example 5: Studio with Custom Environment Variables
resource "seqera_studios" "studio_with_env_vars" {
  auto_start     = false
  compute_env_id = "htaAEef9YYm5DqQrAyeDy"
  configuration = {
    cpu            = 2
    memory         = 8192
    lifespan_hours = 8
    # Studio-specific environment variables (keys must be alphanumeric + underscore, cannot start with number)
    environment = {
      MY_STUDIO_VAR = "testing"
      API_ENDPOINT  = "https://api.example.com"
      DEBUG_MODE    = "true"
    }
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-ride:2025.04.1-0.8"
  description          = "Studio with custom environment variables"
  is_private           = true
  name                 = "studio-with-env"
  workspace_id         = seqera_workspace.my_workspace.id
}


