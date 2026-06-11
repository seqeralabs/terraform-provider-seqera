---
page_title: "Dynamic Resource Labels"
subcategory: "Examples"
description: |-
  Define Seqera resource labels whose values are populated per-run — ${sessionId}, ${workflowId}, ${userName} — and apply them to compute environments and pipelines with Terraform.
---

# Dynamic Resource Labels

Dynamic resource labels extend the standard resource labels functionality by allowing variable values that are populated with unique workflow identifiers at runtime. Where a standard resource label tags your cloud infrastructure with a fixed value, a dynamic resource label uses a variable placeholder that Seqera and Nextflow resolve when a workflow runs. This lets you attribute cloud spend, trace resources, and audit usage at the granularity of an individual run.

This guide shows how to define dynamic resource labels with the `seqera_labels` resource and apply them to your compute environments and pipelines. For the platform-side model, see the [Resource labels overview](https://docs.seqera.io/platform-cloud/resource-labels/overview#dynamic-resource-labels).

## Variable placeholders

Dynamic resource labels use variable placeholders. You can use any of the three placeholders below; Seqera and Nextflow resolve them when a workflow runs. For example, a dynamic resource label `platform-run-id=${workflowId}` becomes `platform-run-id=12345abcde` on the resources a run provisions.

| Placeholder     | Resolves to                                |
| --------------- | ------------------------------------------ |
| `${sessionId}`  | Nextflow session ID                        |
| `${workflowId}` | Platform run ID                            |
| `${userName}`   | Platform username (run launch user)        |

The provider rejects any other `${...}` value at plan time.

## Define a dynamic resource label

To create a dynamic resource label, set `resource = true` and use a variable placeholder in the `value` field:

```terraform
resource "seqera_labels" "session_id" {
  workspace_id = data.seqera_workspace.main.id
  name         = "nextflow-session-id"
  value        = "$${sessionId}"
  resource     = true
  is_default   = true
}
```

~> **Escape the `$` for Terraform.** Terraform interprets `${...}` as its own interpolation syntax. To store the literal value `${sessionId}`, you must double the leading dollar sign: `"$${sessionId}"`. If you write `"${sessionId}"`, Terraform looks for a variable named `sessionId` and the configuration fails to parse. The value stored on the platform and shown in the UI is the single-dollar form, `${sessionId}`.

Set `is_default = true` to apply the label automatically to new resources in the workspace, or omit it to apply the label selectively to specific compute environments and pipelines.

## Apply dynamic resource labels

You apply resource labels — dynamic or standard — to a compute environment or pipeline through its `label_ids` field, which references the `label_id` of each `seqera_labels` resource. You can mix dynamic and standard labels in the same list:

```terraform
resource "seqera_labels" "session_id" {
  workspace_id = data.seqera_workspace.main.id
  name         = "nextflow-session-id"
  value        = "$${sessionId}"
  resource     = true
}

resource "seqera_labels" "run_id" {
  workspace_id = data.seqera_workspace.main.id
  name         = "platform-run-id"
  value        = "$${workflowId}"
  resource     = true
}

resource "seqera_labels" "environment" {
  workspace_id = data.seqera_workspace.main.id
  name         = "environment"
  value        = "production" # a standard (static) value alongside the dynamic ones
  resource     = true
}

resource "seqera_gcp_batch_ce" "main" {
  name         = "gcp-batch"
  workspace_id = data.seqera_workspace.main.id

  label_ids = [
    seqera_labels.session_id.label_id,
    seqera_labels.run_id.label_id,
    seqera_labels.environment.label_id,
  ]

  config = {
    # ... compute environment configuration ...
  }
}
```

The same `label_ids` field is available on `seqera_pipeline` and the other compute environment resources, including `seqera_aws_batch_ce` and `seqera_azure_batch_ce`.

~> **Do not place variable placeholders in `config.labels`.** The `config.labels` map on a compute environment holds cloud-provider tags (for example, Google Cloud resource labels) that Seqera applies when it creates the compute environment. Dynamic resource labels are resolved at workflow submission time, not at compute environment creation time, so a placeholder such as `${sessionId}` is not a valid cloud tag there. The apply may appear to succeed, the resource labels section in the platform UI is left blank, and any run launched on that compute environment fails to start with `Bad Request`. Always model dynamic resource labels as `seqera_labels` resources and attach them through `label_ids`.

## Value format

A resource label value must be either a supported variable placeholder or a standard value. A standard value must:

- Contain 2-39 alphanumeric characters (`a-z`, `A-Z`, `0-9`)
- Use single dashes (`-`) or underscores (`_`) as separators
- Not begin or end with a separator, and not contain consecutive separators

The provider validates the value at plan time, so a malformed value is caught before apply:

```
Error: Invalid Label Value Format

Label value must be 2-39 alphanumeric characters (a-z, A-Z, 0-9) separated by
single dashes (-) or underscores (_), or a dynamic placeholder: ${sessionId},
${workflowId}, or ${userName}.
```

## Prerequisites and limitations

- **Dynamic values require resource labels.** A variable placeholder is only valid when `resource = true`. The provider rejects a `value` on a non-resource label.
- **Dynamic label values cannot be changed in place.** The platform does not allow updating the value of an existing dynamic resource label and returns `Bad Request`. To change a value, replace the resource: `terraform apply -replace=seqera_labels.session_id`.
- **Dynamic resource labels are not supported for Studios.** The platform rejects associating a dynamic resource label with a Studio.
- **Labels apply at submission and execution time.** Nextflow applies dynamic resource labels when a workflow is submitted and runs, not when the compute environment is created.

## Verify the label

After you apply, the platform reports a dynamic resource label with `isDynamic` set to `true`:

```console
$ curl -s -H "Authorization: Bearer $SEQERA_TOKEN" \
    "$SEQERA_API_URL/labels?workspaceId=<id>" | jq '.labels[] | select(.name=="nextflow-session-id")'
{
  "id": 104165001207020,
  "name": "nextflow-session-id",
  "value": "${sessionId}",
  "resource": true,
  "isDefault": true,
  "isDynamic": true,
  "isInterpolated": false
}
```

`isDynamic: true` confirms the platform recognized the placeholder. `isInterpolated: false` is expected: the stored value is the template, which Seqera resolves only when a run is submitted.
