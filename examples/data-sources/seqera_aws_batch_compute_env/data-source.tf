data "seqera_aws_batch_compute_env" "my_awsbatchcomputeenv" {
  attributes = [
    "labels"
  ]
  workspace_id = 10
}