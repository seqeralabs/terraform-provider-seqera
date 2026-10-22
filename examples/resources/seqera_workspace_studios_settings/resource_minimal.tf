# The smallest valid configuration: only the two settings the Platform API
# always needs. container_repository and name_strategy are left unset, so Wave
# uses the platform defaults for where it pushes Studio images and how it names
# them.
resource "seqera_workspace_studios_settings" "minimal" {
  org_id       = seqera_workspace.main.org_id
  workspace_id = seqera_workspace.main.id

  lifespan_hours            = 8
  private_studio_by_default = false
}
