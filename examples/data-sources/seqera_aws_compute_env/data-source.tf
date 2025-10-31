data "seqera_aws_compute_env" "my_awscomputeenv" {
  attributes = [
    "labels"
  ]
  compute_env_id = "...my_compute_env_id..."
  workspace_id   = 7
}