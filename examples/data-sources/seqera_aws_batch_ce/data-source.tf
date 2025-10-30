data "seqera_aws_batch_ce" "my_awsbatchce" {
  attributes = [
    "labels"
  ]
  workspace_id = 5
}