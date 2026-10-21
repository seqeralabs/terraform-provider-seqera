# List the Studio templates configured on the platform. The available templates
# are set per Platform deployment, so they differ between installations.
data "seqera_studio_templates" "all" {
  workspace_id = seqera_workspace.main.id
}

# Pick the recommended image for a tool instead of hard-coding its version
locals {
  recommended_templates = {
    for t in data.seqera_studio_templates.all.templates : t.tool => t.repository
    if t.status == "recommended"
  }
}

resource "seqera_studios" "jupyter" {
  name                 = "my-jupyter-studio"
  compute_env_id       = seqera_compute_env.main.id
  workspace_id         = seqera_workspace.main.id
  data_studio_tool_url = local.recommended_templates["jupyter"]

  configuration = {}
}
