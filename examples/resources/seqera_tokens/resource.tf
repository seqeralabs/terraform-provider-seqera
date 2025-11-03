# Token Resource Examples
#
# Tokens provide API access for programmatic authentication to Seqera Platform.
# The token value (access_key) is only available immediately after creation.

# Example 1: Basic token creation
# Create a token for CI/CD pipeline authentication

resource "seqera_tokens" "ci_pipeline" {
  name = "ci-cd-pipeline-token"
}

# Capture the token value in a sensitive output
# IMPORTANT: The access_key is only available on creation
output "ci_token_value" {
  value     = seqera_tokens.ci_pipeline.access_key
  sensitive = true
}

# Example 2: Token for automation scripts

resource "seqera_tokens" "automation" {
  name = "automation-script-token"
}

# Example 3: Token with descriptive name for team workflows

resource "seqera_tokens" "data_science_team" {
  name = "data-science-team-token"
}