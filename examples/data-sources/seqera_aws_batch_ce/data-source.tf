data "seqera_aws_batch_ce" "my_awsbatchce" {
  attributes = [
    "labels"
  ]
  compute_env_id = "...my_compute_env_id..."
  workspace_id   = 5
}