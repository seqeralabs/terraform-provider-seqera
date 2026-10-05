---
page_title: "AWS Credentials with Workload Identity Federation"
subcategory: "Examples"
description: |-
  End-to-end example of creating a Seqera AWS credential that uses OIDC workload identity federation to assume an IAM role — no access keys stored in Seqera.
---

# AWS Credentials with Workload Identity Federation

This guide creates a `seqera_aws_credential` in `workloadIdentity` mode. Instead of storing an access key, the Seqera Platform acts as an OIDC identity provider: it mints a short-lived token for each operation, and AWS STS exchanges it for temporary credentials through `sts:AssumeRoleWithWebIdentity` on the IAM role you configure.

~> **Note:** Workload identity federation is gated behind the Identity Federation feature flag. On a Platform instance or organization where it is disabled, creating the credential fails at apply time even though the configuration is valid.

## How the trust works

Seqera presents a token with these claims to AWS STS:

| Claim | Value |
| ----- | ----- |
| `iss` | The Platform public address, for example `https://cloud.seqera.io/api` |
| `aud` | `sts.amazonaws.com` |
| `sub` | `org:<ORG_ID>:wsp:<WORKSPACE_ID>:<WORKLOAD>` for workspaces inside an organization, one subject per workload type |

The workload segment of `sub` identifies what is calling AWS: platform actions (compute environment management and run submission), Data Explorer, Studios, and pipeline launches. The IAM role's trust policy conditions on the audience and on these subjects, so a token minted for a different workspace does not match.

AWS rejects the whole exchange unless the role also allows `sts:TagSession` and `sts:SetSourceIdentity`: Seqera's tokens always carry session tags and a source identity, which attribute each AWS call to the acting user in CloudTrail.

You do not need to assemble these values by hand. The [`seqera_aws_credentials_federation_setup`](../data-sources/aws_credentials_federation_setup.md) data source returns the exact issuer, audience and subjects for a workspace.

## Prerequisites

- Identity Federation enabled for your Seqera organization.
- A Seqera workspace you can create credentials in.
- Permission in the AWS account to create an IAM OIDC identity provider and an IAM role.

## Step 1: Register Seqera as an OIDC provider and create the role

The OIDC provider is registered once per AWS account and Seqera installation. The role's trust policy admits every subject the workspace can present, with the same `org:<ORG_ID>:wsp:<WORKSPACE_ID>:*` wildcard the Platform uses in the trust policy it renders, so workload types added later keep working.

```terraform
terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

variable "workspace_id" {
  type        = number
  description = "Seqera workspace ID that will own the credential."
}

data "seqera_aws_credentials_federation_setup" "this" {
  workspace_id = var.workspace_id
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

resource "aws_iam_openid_connect_provider" "seqera" {
  url             = data.seqera_aws_credentials_federation_setup.this.platform_public_address
  client_id_list  = [data.seqera_aws_credentials_federation_setup.this.audience]
  thumbprint_list = [] # populated by AWS for publicly trusted issuers
}

resource "aws_iam_role" "seqera" {
  name = "seqera-workload-identity"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = aws_iam_openid_connect_provider.seqera.arn }
      Action = [
        "sts:AssumeRoleWithWebIdentity",
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
```

Attach to the role the permissions your workloads need, as you would for an access-key or `role` credential. Workload identity changes how Seqera authenticates to AWS, not what the role must be allowed to do.

If the account already has an OIDC provider for this Seqera installation, look it up with the `aws_iam_openid_connect_provider` data source instead of creating a second one.

To pin the exact subjects instead of the wildcard, list `subject_platform_actions`, `subject_data_explorer`, `subject_studios` and `subject_pipeline_launches` under `StringEquals`. The role then rejects any workload type the Platform adds later until you add its subject, and because credential validation uses `subject_platform_actions`, the credential keeps looking valid while the new feature fails.

## Step 2: Define the Seqera credential

```terraform
resource "seqera_aws_credential" "wif" {
  name         = "aws-wif"
  workspace_id = var.workspace_id

  mode            = "workloadIdentity"
  assume_role_arn = aws_iam_role.seqera.arn
}
```

Referencing `aws_iam_role.seqera.arn` orders the role before the credential. In `workloadIdentity` mode the plan rejects `access_key`, `secret_key` and `use_external_id = true`, and requires `assume_role_arn`. These checks are skipped on Platform installations that allow instance credentials, which do not enforce them.

## Step 3: Apply and use

```shell
terraform apply
```

Reference the credential from any resource that accepts `credentials_id`, such as an AWS Batch compute environment:

```terraform
resource "seqera_aws_batch_ce" "example" {
  name           = "aws-batch-example"
  workspace_id   = var.workspace_id
  credentials_id = seqera_aws_credential.wif.credentials_id
  # ... compute env config ...
}
```

The Platform validates the credential after creation by performing a real token exchange. If the trust is not set up correctly, the credential is marked invalid with a message naming the role and subject AWS rejected. Check that the OIDC provider exists in the role's account, that the trust policy matches the subject, and that it allows all three actions.

## Notes

- The mode cannot be changed after creation. Changing `mode` forces replacement of the credential, which gives it a new ID. Compute environments that reference it get the new `credentials_id` in the same plan and are updated in place. Terraform deletes the old credential first, and pipelines running on those compute environments are stopped, so apply the change when nothing is running.
- Workload identity covers the calls Seqera makes to AWS. The Nextflow head job and its tasks still run under the compute environment's instance role, which needs its own permissions.
- The subjects are derived from the workspace that owns the credential. Changing `workspace_id` forces replacement and produces new subjects, so the trust policy must be updated in lockstep.
- The data source's `cloudtrail_session_tag_keys` lists the session tags Seqera sets, for filtering CloudTrail events by acting user.

## Related

- Resource reference: [`seqera_aws_credential`](../resources/aws_credential.md)
- Data source reference: [`seqera_aws_credentials_federation_setup`](../data-sources/aws_credentials_federation_setup.md)
- [Create an OpenID Connect (OIDC) identity provider in IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html)
- [Create a role for OpenID Connect federation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-idp_oidc.html)
