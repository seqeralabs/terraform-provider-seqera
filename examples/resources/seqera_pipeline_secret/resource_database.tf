variable "database_password" {
  description = "Database password for data access"
  type        = string
  sensitive   = true
}

resource "seqera_pipeline_secret" "db_password" {
  name         = "database_password"
  value        = var.database_password
  workspace_id = seqera_workspace.main.id
}
