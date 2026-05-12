# Seqera handles all infra provisioning; you pick the region and (optionally) size.
resource "seqera_managed_compute_ce" "minimal" {
  name         = "seqera-cloud-small"
  workspace_id = data.seqera_workspace.main.id
  region       = "us-east-1"
}
