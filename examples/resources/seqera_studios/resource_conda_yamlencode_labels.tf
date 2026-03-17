resource "seqera_studios" "jupyter_with_conda_labels" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
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
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  description          = "Jupyter studio for data analysis and visualization"
  is_private           = true
  # Reference label IDs from seqera_labels resources
  label_ids = [
    seqera_labels.environment_prod.id,
    seqera_labels.team_datascience.id
  ]
  name         = "jupyter-with-conda-labels"
  spot         = true
  workspace_id = seqera_workspace.main.id
}
