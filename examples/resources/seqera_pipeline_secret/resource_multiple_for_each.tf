locals {
  secrets = {
    "slack_webhook_url" = var.slack_webhook
    "dockerhub_token"   = var.dockerhub_token
    "api_endpoint_key"  = var.api_key
  }
}

resource "seqera_pipeline_secret" "service_secrets" {
  for_each = local.secrets

  name         = each.key
  value        = each.value
  workspace_id = seqera_workspace.main.id
}
