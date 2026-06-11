resource "seqera_labels" "environment" {
  workspace_id = seqera_workspace.main.id
  name         = "environment"
  value        = "production"
  resource     = true
  is_default   = false
}

# Dynamic resource label. The value is interpolated by the Platform at workflow
# submission time. Supported placeholders: $${sessionId}, $${workflowId}, $${userName}.
# Note: the $$ escapes Terraform interpolation so the literal $${sessionId} is sent.
resource "seqera_labels" "session" {
  workspace_id = seqera_workspace.main.id
  name         = "nextflow-session-id"
  value        = "$${sessionId}"
  resource     = true
  is_default   = true
}
