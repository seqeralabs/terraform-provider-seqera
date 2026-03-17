resource "seqera_studios" "jupyter_with_conda_heredoc" {
  auto_start     = false
  compute_env_id = seqera_compute_env.main.id
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
    cpu               = 2
    memory            = 4096
    lifespan_hours    = 8
    # gpu defaults to 0 (disabled)
  }
  data_studio_tool_url = "public.cr.seqera.io/platform/data-studio-jupyter:4.2.5-0.8"
  description          = "Jupyter studio with conda packages defined using heredoc"
  is_private           = true
  name                 = "jupyter-with-conda-heredoc"
  spot                 = true
  workspace_id         = seqera_workspace.main.id
}
