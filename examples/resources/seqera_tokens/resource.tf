resource "seqera_tokens" "ci_pipeline" {
  name = "ci-cd-pipeline-token"
}

# Capture the token value in a sensitive output
# IMPORTANT: The access_key is only available on creation
output "ci_token_value" {
  value     = seqera_tokens.ci_pipeline.access_key
  sensitive = true
}
