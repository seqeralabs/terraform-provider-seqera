# The values AWS needs before a workload-identity credential can be created.
# Resolved for a workspace context; omit workspace_id for the user context.
data "seqera_aws_credentials_federation_setup" "this" {
  workspace_id = seqera_workspace.my_workspace.id
}

locals {
  # AWS names the provider, and prefixes its condition keys, with the issuer
  # without the scheme or a trailing slash, as the Platform does.
  issuer_host = trimsuffix(
    replace(data.seqera_aws_credentials_federation_setup.this.platform_public_address, "/^https?:///", ""),
    "/",
  )

  # Every subject the workspace presents, one per workload type: the same
  # wildcard the Platform uses in the trust policy it renders.
  subject_pattern = "${trimsuffix(data.seqera_aws_credentials_federation_setup.this.subject_platform_actions, ":platform")}:*"
}

# Register Seqera as an OIDC identity provider in your AWS account.
resource "aws_iam_openid_connect_provider" "seqera" {
  url             = data.seqera_aws_credentials_federation_setup.this.platform_public_address
  client_id_list  = [data.seqera_aws_credentials_federation_setup.this.audience]
  thumbprint_list = [] # populated by AWS for publicly trusted issuers
}

# Trust every subject Seqera mints for this workspace, one per workload type,
# with the wildcard the Platform's own trust policy uses, so workload types
# added later keep working.
resource "aws_iam_role" "seqera" {
  name = "seqera-platform"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = aws_iam_openid_connect_provider.seqera.arn }
      Action = [
        "sts:AssumeRoleWithWebIdentity",
        # Seqera's tokens always carry session tags and a source identity;
        # AWS rejects the whole exchange for a role that does not allow them.
        "sts:TagSession",
        "sts:SetSourceIdentity",
      ]
      Condition = {
        StringEquals = {
          "${local.issuer_host}:aud" = data.seqera_aws_credentials_federation_setup.this.audience
        }
        StringLike = {
          "${local.issuer_host}:sub" = local.subject_pattern
        }
      }
    }]
  })
}

# Feed the role back into the credential that assumes it.
resource "seqera_aws_credential" "workload_identity" {
  name            = "aws-workload-identity"
  workspace_id    = seqera_workspace.my_workspace.id
  assume_role_arn = aws_iam_role.seqera.arn
}

# The session tag keys Seqera sets on the assumed role, for CloudTrail queries.
output "cloudtrail_session_tag_keys" {
  value = data.seqera_aws_credentials_federation_setup.this.cloudtrail_session_tag_keys
}

# Every value as returned by the API, including any label not mapped above.
output "all_setup_values" {
  value = data.seqera_aws_credentials_federation_setup.this.setup_values
}
