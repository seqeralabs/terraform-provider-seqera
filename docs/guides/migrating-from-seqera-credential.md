---
page_title: "Migrating from seqera_credential to typed credential resources"
subcategory: "Upgrade Guides"
description: |-
  Step-by-step guide for migrating existing seqera_credential resources to the per-provider typed credentials (seqera_aws_credential, seqera_bitbucket_credential, seqera_github_credential, etc.) without destroying the underlying Seqera Platform credential.
---

# Migrating from `seqera_credential` to typed credential resources

The provider exposes both a generic `seqera_credential` resource (one resource type, a `keys` block per provider) and per-provider typed resources (`seqera_aws_credential`, `seqera_github_credential`, `seqera_bitbucket_credential`, …). The typed resources have provider-specific schemas, plan-time validators, and clearer documentation. New configurations should use them.

This guide walks through migrating an existing `seqera_credential` to the matching typed resource without destroying the underlying credential on Seqera Platform.

## Why migrate

- Provider-specific plan-time validation (e.g. mutually-exclusive auth fields, required-pair checks).
- Clearer field names — flat top-level attributes instead of a nested `keys.<provider>` block.
- Per-provider documentation pages with field-level guidance.
- Generic `seqera_credential` will eventually be deprecated in favour of the typed resources.

## Approach

Migration uses Terraform's import/removal flow rather than `moved {}` blocks. `moved {}` only works between resources of the same type (or when the destination resource declares an explicit `MoveState` for the source type, which the credential resources do not). What you can do today:

1. **Terraform 1.7+ — single-plan migration** using `import {}` and `removed {}` blocks. Recommended.
2. **Terraform 1.5–1.6** — `terraform state rm` followed by `terraform import` (or an `import {}` block). Same end state, two commands.

The underlying Seqera Platform credential is never touched — only Terraform state is rewritten.

## What you need

- The credential's `credentials_id` (from `terraform state show seqera_credential.<name>`, the `credentials_id` attribute).
- The `workspace_id` it lives in (also visible in state, or in your existing config).
- The credential's secret values (token/password/key data). These are write-only — the provider never reads them back from the API, so the typed resource will need them re-supplied in config (typically via the same `var.…` you already use).

## Provider mapping

The `keys.<provider>` sub-block in `seqera_credential` maps to a typed resource as follows:

| `seqera_credential.keys` block | Typed resource | Notable field renames |
| --- | --- | --- |
| `aws` | `seqera_aws_credential` | `access_key`, `secret_key`, `assume_role_arn` |
| `azure` / `azure_cloud` / `azure_entra` | `seqera_azure_credential` | `batch_name`, `storage_name`, `batch_key`, `storage_key`, `tenant_id`, `client_id`, `client_secret` |
| `bitbucket` | `seqera_bitbucket_credential` | `username`, `password`, `token` |
| `codecommit` | `seqera_codecommit_credential` | `username` → `access_key`, `password` → `secret_key` |
| `container_reg` | `seqera_container_registry_credential` | `username`, `password`, `registry` |
| `gitea` | `seqera_gitea_credential` | `username`, `password` |
| `github` | `seqera_github_credential` | `password` → `access_token` |
| `gitlab` | `seqera_gitlab_credential` | `username`, `token` |
| `google` | `seqera_google_credential` | `data` (service account JSON) |
| `k8s` | `seqera_kubernetes_credential` | `data` (kubeconfig) |
| `ssh` | `seqera_ssh_credential` | `private_key`, `passphrase` |
| `tw_agent` | `seqera_tower_agent_credential` | `connection_id`, `work_dir` |

In all cases the top-level fields (`name`, `workspace_id`, `base_url`, `provider_type`) keep their names.

## Worked example: Bitbucket

### Before

```terraform
resource "seqera_credential" "bitbucket" {
  name         = "bitbucket-main"
  workspace_id = seqera_workspace.main.id
  base_url     = "https://bitbucket.org/seqeralabs"

  keys = {
    bitbucket = {
      username = var.bitbucket_username
      password = var.bitbucket_password
    }
  }
}
```

### Step 1 — capture the IDs

```shell
terraform state show seqera_credential.bitbucket | grep -E 'credentials_id|workspace_id'
```

Note the values; you'll need them in step 2.

### Step 2 — rewrite the resource and add migration blocks (Terraform 1.7+)

```terraform
resource "seqera_bitbucket_credential" "bitbucket" {
  name         = "bitbucket-main"
  workspace_id = seqera_workspace.main.id
  base_url     = "https://bitbucket.org/seqeralabs"

  username = var.bitbucket_username
  password = var.bitbucket_password
}

import {
  to = seqera_bitbucket_credential.bitbucket
  id = jsonencode({
    credentials_id = "abc123XYZ"   # from step 1
    workspace_id   = 1234567890    # from step 1
  })
}

removed {
  from = seqera_credential.bitbucket
  lifecycle {
    destroy = false   # important — keeps the Platform credential alive
  }
}
```

### Step 3 — apply

```shell
terraform plan
terraform apply
```

Terraform imports the credential under the new resource address and drops the old one from state without calling Delete. Confirm with:

```shell
terraform state list | grep credential
```

You should see `seqera_bitbucket_credential.bitbucket` and no `seqera_credential.bitbucket`.

### Step 4 — clean up

Once the plan is clean (no diff), remove the `import {}` and `removed {}` blocks. They are one-shot — leaving them in is harmless but adds noise.

## Older Terraform (1.5–1.6)

Same end state, two steps:

```shell
# 1. Drop the old resource from state without touching the Platform credential
terraform state rm seqera_credential.bitbucket

# 2. Import under the new typed resource address
terraform import 'seqera_bitbucket_credential.bitbucket' \
  '{"credentials_id": "abc123XYZ", "workspace_id": 1234567890}'
```

Then run `terraform plan` and confirm no diff. The typed `resource "seqera_bitbucket_credential" "bitbucket" { … }` block must already be in your config before the import.

## Worked example: GitHub

### Before

```terraform
resource "seqera_credential" "github" {
  name         = "github-main"
  workspace_id = seqera_workspace.main.id
  base_url     = "https://github.com/seqeralabs"

  keys = {
    github = {
      username = var.github_username
      password = var.github_token   # PAT goes in `password`
    }
  }
}
```

### After

```terraform
resource "seqera_github_credential" "github" {
  name         = "github-main"
  workspace_id = seqera_workspace.main.id
  base_url     = "https://github.com/seqeralabs"

  username     = var.github_username
  access_token = var.github_token   # renamed from `password`
}

import {
  to = seqera_github_credential.github
  id = jsonencode({
    credentials_id = "..."
    workspace_id   = 0
  })
}

removed {
  from = seqera_credential.github
  lifecycle { destroy = false }
}
```

The only field-level change: the PAT moves from `keys.github.password` to a top-level `access_token`.

## Worked example: AWS

### Before

```terraform
resource "seqera_credential" "aws" {
  name         = "aws-main"
  workspace_id = seqera_workspace.main.id

  keys = {
    aws = {
      accessKey     = var.aws_access_key
      secretKey     = var.aws_secret_key
      assumeRoleArn = "arn:aws:iam::123456789012:role/SeqeraRole"
    }
  }
}
```

### After

```terraform
resource "seqera_aws_credential" "aws" {
  name         = "aws-main"
  workspace_id = seqera_workspace.main.id

  access_key      = var.aws_access_key
  secret_key      = var.aws_secret_key
  assume_role_arn = "arn:aws:iam::123456789012:role/SeqeraRole"
}

import {
  to = seqera_aws_credential.aws
  id = jsonencode({
    credentials_id = "..."
    workspace_id   = 0
  })
}

removed {
  from = seqera_credential.aws
  lifecycle { destroy = false }
}
```

Field renames: `accessKey` → `access_key`, `secretKey` → `secret_key`, `assumeRoleArn` → `assume_role_arn`.

## Verifying the migration

After apply:

- `terraform plan` should report **no changes**.
- The Seqera Platform UI should show the same credential record (same ID, same name, same workspace).
- Any compute environments / pipelines / actions referencing the credential by ID continue to work without re-creation, because the underlying Platform record is unchanged.

## Common pitfalls

- **Forgot `lifecycle { destroy = false }` in the `removed {}` block.** Without it, Terraform will Delete the credential — which destroys the Platform record, including any compute environment associations. Always include it.
- **`workspace_id` is `0` in the import ID.** The `0` placeholder in the snippets is just illustrative — replace it with the real numeric workspace ID. A `0` value will fail the import.
- **Secret values out of sync after import.** Write-only fields (`password`, `token`, `access_token`, `secret_key`, etc.) are never read back from the API. After import, the config's value is what gets sent on the next Update; if the value differs from what's in the platform you'll silently rotate the secret on the next apply. Re-supply the same secret you used originally.
- **Leftover `import {}` / `removed {}` blocks.** They're one-shot. Once apply is clean, remove them — otherwise every plan re-evaluates them (no-op but noisy).
