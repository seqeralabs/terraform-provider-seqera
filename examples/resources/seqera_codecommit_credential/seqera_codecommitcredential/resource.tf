# Seqera CodeCommit Credentials Examples
#
# CodeCommit credentials store AWS authentication information for accessing
# AWS CodeCommit repositories within the Seqera Platform workflows.
#
# SECURITY BEST PRACTICES:
# - Never hardcode credentials in Terraform files
# - Use Terraform variables marked as sensitive
# - Store actual credentials in secure secret management systems
# - Prefer IAM roles with temporary credentials when possible
# - Restrict IAM permissions to minimum required (principle of least privilege)
# - Consider using AWS Secrets Manager or AWS Systems Manager Parameter Store
# - Regularly rotate access keys

# Variable declarations for sensitive CodeCommit credentials
variable "codecommit_access_key" {
  description = "AWS access key for CodeCommit"
  type        = string
  sensitive   = true
}

variable "codecommit_secret_key" {
  description = "AWS secret key for CodeCommit"
  type        = string
  sensitive   = true
}

# =============================================================================
# Example 1: Basic CodeCommit Credentials (Recommended)
# =============================================================================
# Basic configuration with AWS IAM user credentials for CodeCommit access.

resource "seqera_codecommit_credential" "basic" {
  name         = "codecommit-main"
  workspace_id = seqera_workspace.main.id

  access_key = var.codecommit_access_key
  secret_key = var.codecommit_secret_key
}

# =============================================================================
# Example 2: CodeCommit Credentials with Specific Region Base URL (Recommended)
# =============================================================================
# Specify base_url to associate credentials with a specific AWS region.
# This is recommended for better performance and to avoid cross-region data transfer.

resource "seqera_codecommit_credential" "with_region" {
  name         = "codecommit-us-east-1"
  workspace_id = seqera_workspace.main.id

  access_key = var.codecommit_access_key
  secret_key = var.codecommit_secret_key
  base_url   = "https://git-codecommit.us-east-1.amazonaws.com"
}

# =============================================================================
# Example 3: Multiple CodeCommit Credentials for Different Regions
# =============================================================================

locals {
  codecommit_regions = {
    "us-east-1" = {
      access_key = var.codecommit_us_east_1_access_key
      secret_key = var.codecommit_us_east_1_secret_key
      base_url   = "https://git-codecommit.us-east-1.amazonaws.com"
    }
    "eu-west-1" = {
      access_key = var.codecommit_eu_west_1_access_key
      secret_key = var.codecommit_eu_west_1_secret_key
      base_url   = "https://git-codecommit.eu-west-1.amazonaws.com"
    }
    "ap-southeast-1" = {
      access_key = var.codecommit_ap_southeast_1_access_key
      secret_key = var.codecommit_ap_southeast_1_secret_key
      base_url   = "https://git-codecommit.ap-southeast-1.amazonaws.com"
    }
  }
}

# Note: You would need to declare the corresponding variables:
# variable "codecommit_us_east_1_access_key" { type = string; sensitive = true }
# variable "codecommit_us_east_1_secret_key" { type = string; sensitive = true }
# variable "codecommit_eu_west_1_access_key" { type = string; sensitive = true }
# variable "codecommit_eu_west_1_secret_key" { type = string; sensitive = true }
# variable "codecommit_ap_southeast_1_access_key" { type = string; sensitive = true }
# variable "codecommit_ap_southeast_1_secret_key" { type = string; sensitive = true }

resource "seqera_codecommit_credential" "multi_region" {
  for_each = local.codecommit_regions

  name         = "codecommit-${each.key}"
  workspace_id = seqera_workspace.main.id
  access_key   = each.value.access_key
  secret_key   = each.value.secret_key
  base_url     = each.value.base_url
}

# =============================================================================
# Example 4: Creating IAM User for CodeCommit Access
# =============================================================================
# To create IAM credentials for CodeCommit access:
#
# 1. Create IAM User with AWS CLI:
#    aws iam create-user --user-name codecommit-seqera
#
# 2. Attach CodeCommit access policy:
#    aws iam attach-user-policy \
#      --user-name codecommit-seqera \
#      --policy-arn arn:aws:iam::aws:policy/AWSCodeCommitPowerUser
#
# 3. Create access keys:
#    aws iam create-access-key --user-name codecommit-seqera
#
# 4. Save the AccessKeyId and SecretAccessKey from the output
#
# 5. Use the credentials in your Terraform configuration:
#    - Access Key: Use as access_key
#    - Secret Key: Use as secret_key
#
# SECURITY NOTE: Use least privilege principle. Instead of AWSCodeCommitPowerUser,
# consider creating a custom policy with only the required permissions:
#
# {
#   "Version": "2012-10-17",
#   "Statement": [
#     {
#       "Effect": "Allow",
#       "Action": [
#         "codecommit:GitPull",
#         "codecommit:GitPush"
#       ],
#       "Resource": "arn:aws:codecommit:REGION:ACCOUNT:REPOSITORY_NAME"
#     }
#   ]
# }

# =============================================================================
# Example 5: Using CodeCommit Credentials with Pipelines
# =============================================================================

resource "seqera_codecommit_credential" "pipeline_creds" {
  name         = "codecommit-pipelines"
  workspace_id = seqera_workspace.main.id
  access_key   = var.codecommit_access_key
  secret_key   = var.codecommit_secret_key
  base_url     = "https://git-codecommit.us-east-1.amazonaws.com"
}

resource "seqera_pipeline" "from_codecommit" {
  name         = "my-pipeline"
  workspace_id = seqera_workspace.main.id

  # Reference the CodeCommit repository
  repository = "https://git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo"

  # Use the CodeCommit credentials
  credentials_id = seqera_codecommit_credential.pipeline_creds.credentials_id
}

# =============================================================================
# Example 6: Using AWS Secrets Manager for Credential Storage
# =============================================================================

# Retrieve CodeCommit credentials from AWS Secrets Manager
data "aws_secretsmanager_secret_version" "codecommit_creds" {
  secret_id = "seqera/codecommit/credentials"
}

locals {
  codecommit_secret = jsondecode(data.aws_secretsmanager_secret_version.codecommit_creds.secret_string)
}

resource "seqera_codecommit_credential" "from_secrets_manager" {
  name         = "codecommit-secure"
  workspace_id = seqera_workspace.main.id
  access_key   = local.codecommit_secret.access_key
  secret_key   = local.codecommit_secret.secret_key
  base_url     = local.codecommit_secret.base_url
}

# To store credentials in AWS Secrets Manager:
# aws secretsmanager create-secret \
#   --name seqera/codecommit/credentials \
#   --secret-string '{"access_key":"AKIAIOSFODNN7EXAMPLE","secret_key":"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY","base_url":"https://git-codecommit.us-east-1.amazonaws.com"}'

# =============================================================================
# Example 7: CodeCommit Credentials with Git-Remote-CodeCommit (GRC)
# =============================================================================
# When using git-remote-codecommit, you may need specific IAM permissions.
# Ensure your IAM user has:
# - codecommit:GitPull
# - codecommit:GitPush
#
# The base_url should point to the CodeCommit service endpoint in your region:

resource "seqera_codecommit_credential" "with_grc" {
  name         = "codecommit-grc"
  workspace_id = seqera_workspace.main.id
  access_key   = var.codecommit_access_key
  secret_key   = var.codecommit_secret_key
  base_url     = "https://git-codecommit.eu-central-1.amazonaws.com"
}
