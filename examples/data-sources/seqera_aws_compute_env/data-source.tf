data "seqera_aws_compute_env" "my_awscomputeenv" {
  attributes = [
    "labels"
  ]
  workspace_id = 7
}