resource "seqera_primary_compute_env" "default" {
  workspace_id   = seqera_workspace.main.id
  compute_env_id = seqera_compute_env.main.id
}
