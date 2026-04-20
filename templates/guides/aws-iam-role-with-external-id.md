---
page_title: "AWS Credentials with IAM Role and External ID"
subcategory: "Examples"
description: |-
  End-to-end example of creating a Seqera AWS credential that assumes a cross-account IAM role using a platform-generated external ID.
---

# AWS Credentials with IAM Role and External ID

This guide creates a `seqera_aws_credential` that assumes a cross-account IAM role using a platform-generated external ID, then wires that external ID into the role's trust policy — either manually or via the `hashicorp/aws` provider in the same Terraform configuration.

~> **Note:** On Seqera Cloud (`api.cloud.seqera.io`), role-based credentials without an external ID are rejected by the platform. The external-ID flow shown here is the supported path and is also the most secure, since the platform-issued identifier is required in the trust policy condition and prevents confused-deputy attacks.

## Prerequisites

- A Seqera Platform workspace you can create credentials in.
- Permission to create or update an IAM role in the AWS account Seqera will access.

## Step 1: Define the credential

```terraform
terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
  }
}

variable "workspace_id" {
  type        = number
  description = "Seqera workspace ID that will own the credential."
}

variable "assume_role_arn" {
  type        = string
  description = "ARN of the IAM role Seqera will assume."
}

resource "seqera_aws_credential" "cross_account" {
  name         = "aws-cross-account"
  workspace_id = var.workspace_id

  assume_role_arn = var.assume_role_arn
  mode            = "role"
  use_external_id = true
}
```

`mode = "role"` configures the credential for role assumption without access keys. `use_external_id = true` requests a tenant-scoped external ID, which the platform generates and returns on the created credential.

## Step 2: Surface the generated external ID

```terraform
output "seqera_external_id" {
  description = "External ID for the AWS IAM role trust policy. Use this in the role's sts:AssumeRole condition."
  value       = seqera_aws_credential.cross_account.external_id
  sensitive   = true
}
```

`external_id` is sensitive, so Terraform redacts it from plan and apply output. Retrieve the raw value with `terraform output -raw seqera_external_id` when you need to paste it into the AWS trust policy.

## Step 3: Update the IAM role trust policy

In the AWS account that owns the role referenced by `assume_role_arn`, edit the role's trust policy to require the external ID. A minimal policy looks like:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::161471496260:role/SeqeraPlatformCloudAccessRole"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "<value from seqera_external_id output>"
        }
      }
    },
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::161471496260:role/SeqeraPlatformCloudAccessRole"
      },
      "Action": "sts:TagSession"
    }
  ]
}
```

`arn:aws:iam::161471496260:role/SeqeraPlatformCloudAccessRole` is the Seqera Platform cloud role that performs the `sts:AssumeRole` and `sts:TagSession` calls on your behalf. Substitute the external ID from step 2 for `<value from seqera_external_id output>`. The authoritative trust policy reference is in the Seqera docs: [Role-based trust policy example (Seqera Cloud)](https://docs.seqera.io/platform-cloud/compute-envs/aws-batch#role-based-trust-policy-example-seqera-cloud).

## Step 4: Apply, verify, and use

```shell
terraform apply
terraform output -raw seqera_external_id
# Copy the printed value into the IAM role trust policy so Seqera can assume the role.
```

Once the trust policy is live, reference the credential from any resource that accepts `credentials_id`, such as a compute environment:

```terraform
resource "seqera_aws_batch_ce" "example" {
  name           = "aws-batch-example"
  workspace_id   = var.workspace_id
  credentials_id = seqera_aws_credential.cross_account.credentials_id
  # ... compute env config ...
}
```

## Managing the IAM role with the AWS provider

The manual flow above requires copying the external ID into the AWS console. To manage the IAM role from the same Terraform configuration, pair the `seqera` provider with `hashicorp/aws`.

Before you begin, pick the IAM role name you want Seqera to assume. The order of operations is:

1. Build the target role ARN as a `local` from the AWS account ID and the chosen role name. The Seqera credential references this local, not the `aws_iam_role` resource.
2. Terraform creates `seqera_aws_credential`, which stores the target ARN and returns the platform-generated `external_id`.
3. Terraform creates `aws_iam_role` with a trust policy that embeds the `external_id` in the `sts:ExternalId` condition.
4. Downstream resources that trigger `sts:AssumeRole` (compute environments and similar) set `depends_on = [aws_iam_role.seqera_access]` so the trust policy is live before Seqera uses the credential.

```terraform
terraform {
  required_providers {
    seqera = {
      source = "seqeralabs/seqera"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

variable "workspace_id" {
  type = number
}

variable "iam_role_name" {
  type        = string
  default     = "SeqeraAccess"
  description = "Name of the IAM role Seqera will assume."
}

data "aws_caller_identity" "current" {}

locals {
  iam_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.iam_role_name}"
}

resource "seqera_aws_credential" "cross_account" {
  name            = "aws-cross-account"
  workspace_id    = var.workspace_id
  assume_role_arn = local.iam_role_arn
  mode            = "role"
  use_external_id = true
}

resource "aws_iam_role" "seqera_access" {
  name = var.iam_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::161471496260:role/SeqeraPlatformCloudAccessRole"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "sts:ExternalId" = seqera_aws_credential.cross_account.external_id
          }
        }
      },
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::161471496260:role/SeqeraPlatformCloudAccessRole"
        }
        Action = "sts:TagSession"
      }
    ]
  })
}

# Attach Seqera's required permission policy here — see the Required
# Platform IAM permissions reference linked below for the policy document.
#
# resource "aws_iam_role_policy" "seqera_permissions" {
#   role   = aws_iam_role.seqera_access.name
#   name   = "SeqeraPlatformPermissions"
#   policy = file("${path.module}/seqera-platform-permissions.json")
# }
```

### Variant — attach to an existing IAM role

If the IAM role is already managed out of band — for example in a separate AWS account, or provisioned by another team — skip the `aws_iam_role` resource and look up the existing role by name. The trust policy still needs the `sts:AssumeRole` and `sts:TagSession` statements referencing the generated external ID; either have the role owner apply them, or run a second `terraform apply` against the `aws` provider once the external ID is known.

```terraform
variable "existing_role_name" {
  type        = string
  description = "Name of an IAM role that already exists and trusts Seqera."
}

data "aws_iam_role" "existing" {
  name = var.existing_role_name
}

resource "seqera_aws_credential" "cross_account" {
  name            = "aws-cross-account"
  workspace_id    = var.workspace_id
  assume_role_arn = data.aws_iam_role.existing.arn
  mode            = "role"
  use_external_id = true
}

output "seqera_external_id" {
  value     = seqera_aws_credential.cross_account.external_id
  sensitive = true
}
```

Run `terraform apply`, then share `seqera_external_id` with the role owner so they can add it to the trust policy conditions shown above.

## Notes

- The `external_id` is generated once at credential creation time and cannot be rotated in place. To rotate it, delete and recreate the credential.
- Changing `var.iam_role_name` in the combined flow replaces both the credential and the role together, which issues a new external ID. The plan output will reflect this before apply.
- `workspace_id` is the numeric workspace ID, not the slug. Retrieve it from the workspace URL or the `seqera_workspace` data source.
- The credential is tied to the workspace it was created in. Changing `workspace_id` forces replacement.

## Related

- Resource reference: [`seqera_aws_credential`](../resources/aws_credential.md)
- [Role-based trust policy example (Seqera Cloud)](https://docs.seqera.io/platform-cloud/compute-envs/aws-batch#role-based-trust-policy-example-seqera-cloud) — authoritative trust policy reference used above.
- [Required Platform IAM permissions](https://docs.seqera.io/platform-cloud/compute-envs/aws-batch#required-platform-iam-permissions) — authoritative permission set the assumed role must have.
