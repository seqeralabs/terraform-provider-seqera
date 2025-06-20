data "seqera_compute_env" "my_computeenv" {
  attributes = [
    "labels"
  ]
  compute_env_id = "...my_compute_env_id..."
  workspace_id   = 1
}