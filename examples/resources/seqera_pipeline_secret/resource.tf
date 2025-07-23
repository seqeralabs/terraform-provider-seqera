resource "seqera_pipeline_secret" "my_pipelinesecret" {
  name         = "database_password"
  value        = "super-secret-password-123"
  workspace_id = 8
}
