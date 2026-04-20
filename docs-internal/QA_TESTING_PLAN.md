# Seqera Terraform Provider - Manual QA & Testing Plan

This document provides a comprehensive manual QA and testing plan for all resources and data sources in the Seqera Terraform Provider.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Testing Environment Setup](#testing-environment-setup)
3. [Testing Methodology](#testing-methodology)
4. [Resources Testing](#resources-testing)
5. [Data Sources Testing](#data-sources-testing)
6. [Integration Testing Scenarios](#integration-testing-scenarios)
7. [Error Handling & Edge Cases](#error-handling--edge-cases)
8. [Test Results Template](#test-results-template)

---

## Prerequisites

### Required Access

- [ ] Seqera Platform account with admin privileges
- [ ] At least one organization with owner role
- [ ] API access token with full permissions
- [ ] Cloud provider credentials (AWS/Azure/GCP) for compute environment testing

### Required Tools

- [ ] Terraform CLI (v1.0+)
- [ ] Go (for building the provider locally)
- [ ] Access to cloud provider consoles for verification

### Environment Variables

```bash
export SEQERA_API_TOKEN="your-api-token"
export SEQERA_API_URL="https://api.cloud.seqera.io"  # or your private deployment
```

---

## Testing Environment Setup

### 1. Build the Provider Locally

```bash
cd terraform-provider-seqera
go build -o terraform-provider-seqera
```

### 2. Configure Terraform for Local Provider

Create/update `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/seqeralabs/seqera" = "/path/to/terraform-provider-seqera"
  }
  direct {}
}
```

### 3. Create Test Directory

```bash
mkdir -p examples/tests/qa
cd examples/tests/qa
```

---

## Testing Methodology

For each resource/data source, perform the following test phases:

### Phase 1: Create (C)

- Run `terraform plan` - verify plan shows expected changes
- Run `terraform apply` - verify resource is created
- Verify resource exists in Seqera Platform UI/API

### Phase 2: Read (R)

- Run `terraform plan` again - should show no changes
- Run `terraform refresh` - state should remain consistent

### Phase 3: Update (U)

- Modify a mutable attribute in the configuration
- Run `terraform plan` - verify update is detected
- Run `terraform apply` - verify update succeeds
- Verify changes in Seqera Platform UI/API

### Phase 4: Delete (D)

- Run `terraform destroy`
- Verify resource is removed from Seqera Platform

### Phase 5: Import (I)

- Create resource manually in Seqera Platform
- Run `terraform import` with correct ID format
- Run `terraform plan` - should show no changes (or expected drift)

---

## Resources Testing

### 1. Organization Resources

#### 1.1 seqera_orgs

**Description:** Manages Seqera organizations

**Test Configuration:**

```hcl
resource "seqera_orgs" "test" {
  name        = "qa-test-org"
  full_name   = "QA Test Organization"
  description = "Organization for QA testing"
}
```

| Test Case          | Steps                                        | Expected Result         | Pass/Fail |
| ------------------ | -------------------------------------------- | ----------------------- | --------- |
| Create org         | Apply config                                 | Org created in Platform |           |
| Read org           | Plan after create                            | No changes              |           |
| Update description | Change description, apply                    | Description updated     |           |
| Delete org         | Destroy                                      | Org removed             |           |
| Import org         | `terraform import seqera_orgs.test <org_id>` | State imported          |           |

---

#### 1.2 seqera_organization_member

**Description:** Manages organization membership

**Prerequisites:** Existing organization, user email that exists in Seqera

**Test Configuration:**

```hcl
resource "seqera_organization_member" "test" {
  org_id = seqera_orgs.test.id
  email  = "testuser@example.com"
  role   = "member"
}
```

| Test Case                   | Steps                                                               | Expected Result  | Pass/Fail |
| --------------------------- | ------------------------------------------------------------------- | ---------------- | --------- |
| Create member with email    | Apply with email                                                    | Member added     |           |
| Read member                 | Plan after create                                                   | No changes       |           |
| Update role to owner        | Change role, apply                                                  | Role updated     |           |
| Update role to collaborator | Change role, apply                                                  | Role updated     |           |
| Delete member               | Destroy                                                             | Member removed   |           |
| Import member               | `terraform import seqera_organization_member.test <org_id>/<email>` | State imported   |           |
| Create duplicate (error)    | Apply with existing member                                          | 409 error        |           |
| Invalid role (error)        | Use role "invalid"                                                  | Validation error |           |

---

### 2. Workspace Resources

#### 2.1 seqera_workspace

**Description:** Manages workspaces within an organization

**Test Configuration:**

```hcl
resource "seqera_workspace" "test" {
  org_id      = seqera_orgs.test.id
  name        = "qa-test-workspace"
  full_name   = "QA Test Workspace"
  description = "Workspace for QA testing"
  visibility  = "PRIVATE"
}
```

| Test Case                | Steps                                                   | Expected Result     | Pass/Fail |
| ------------------------ | ------------------------------------------------------- | ------------------- | --------- |
| Create private workspace | Apply config                                            | Workspace created   |           |
| Read workspace           | Plan after create                                       | No changes          |           |
| Update description       | Change description, apply                               | Description updated |           |
| Change visibility        | Change to PUBLIC, apply                                 | Visibility updated  |           |
| Delete workspace         | Destroy                                                 | Workspace removed   |           |
| Import workspace         | `terraform import seqera_workspace.test <workspace_id>` | State imported      |           |

---

#### 2.2 seqera_workspace_participant

**Description:** Manages workspace participants

**Prerequisites:** Existing workspace, organization member

**Test Configuration:**

```hcl
resource "seqera_workspace_participant" "test" {
  org_id       = seqera_orgs.test.id
  workspace_id = seqera_workspace.test.id
  email        = "testuser@example.com"
  role         = "view"
}
```

| Test Case                           | Steps                                                                                | Expected Result                  | Pass/Fail |
| ----------------------------------- | ------------------------------------------------------------------------------------ | -------------------------------- | --------- |
| Create with email                   | Apply with email                                                                     | Participant added with view role |           |
| Create with member_id               | Apply with member_id                                                                 | Participant added                |           |
| Read participant                    | Plan after create                                                                    | No changes                       |           |
| Update role to launch               | Change role, apply                                                                   | Role updated                     |           |
| Update role to maintain             | Change role, apply                                                                   | Role updated                     |           |
| Update role to admin                | Change role, apply                                                                   | Role updated                     |           |
| Update role to owner                | Change role, apply                                                                   | Role updated                     |           |
| Delete participant                  | Destroy                                                                              | Participant removed              |           |
| Import participant                  | `terraform import seqera_workspace_participant.test <org_id>/<workspace_id>/<email>` | State imported                   |           |
| Both email and member_id (error)    | Specify both                                                                         | Validation error                 |           |
| Neither email nor member_id (error) | Specify neither                                                                      | Validation error                 |           |

---

### 3. Team Resources

#### 3.1 seqera_teams

**Description:** Manages teams within an organization

**Test Configuration:**

```hcl
resource "seqera_teams" "test" {
  org_id      = seqera_orgs.test.id
  name        = "qa-test-team"
  description = "Team for QA testing"
}
```

| Test Case          | Steps                                          | Expected Result     | Pass/Fail |
| ------------------ | ---------------------------------------------- | ------------------- | --------- |
| Create team        | Apply config                                   | Team created        |           |
| Read team          | Plan after create                              | No changes          |           |
| Update description | Change description, apply                      | Description updated |           |
| Delete team        | Destroy                                        | Team removed        |           |
| Import team        | `terraform import seqera_teams.test <team_id>` | State imported      |           |

---

#### 3.2 seqera_team_member

**Description:** Manages team membership

**Prerequisites:** Existing team, organization member

**Test Configuration:**

```hcl
resource "seqera_team_member" "test" {
  org_id  = seqera_orgs.test.id
  team_id = seqera_teams.test.id
  email   = "testuser@example.com"
}
```

| Test Case             | Steps                                                                 | Expected Result                         | Pass/Fail |
| --------------------- | --------------------------------------------------------------------- | --------------------------------------- | --------- |
| Create with email     | Apply with email                                                      | Member added to team                    |           |
| Create with member_id | Apply with member_id                                                  | Member added to team                    |           |
| Read team member      | Plan after create                                                     | No changes                              |           |
| Delete team member    | Destroy                                                               | Member removed from team                |           |
| Import team member    | `terraform import seqera_team_member.test <org_id>/<team_id>/<email>` | State imported                          |           |
| Update (error)        | Change email                                                          | Forces replacement (no in-place update) |           |

---

### 4. Credential Resources

#### 4.1 seqera_credential (Generic)

**Description:** Generic credential resource

**Test Configuration:**

```hcl
resource "seqera_credential" "test" {
  workspace_id = seqera_workspace.test.id
  name         = "qa-test-credential"
  description  = "Credential for QA testing"
  # Provider-specific fields...
}
```

| Test Case          | Steps                     | Expected Result     | Pass/Fail |
| ------------------ | ------------------------- | ------------------- | --------- |
| Create credential  | Apply config              | Credential created  |           |
| Read credential    | Plan after create         | No changes          |           |
| Update description | Change description, apply | Description updated |           |
| Delete credential  | Destroy                   | Credential removed  |           |

---

#### 4.2 seqera_aws_credential

**Test Configuration:**

```hcl
resource "seqera_aws_credential" "test" {
  workspace_id = seqera_workspace.test.id
  name         = "qa-aws-credential"
  access_key   = "AKIAIOSFODNN7EXAMPLE"
  secret_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
```

| Test Case                  | Steps             | Expected Result                     | Pass/Fail |
| -------------------------- | ----------------- | ----------------------------------- | --------- |
| Create AWS credential      | Apply config      | Credential created                  |           |
| Verify secret not in state | Check state file  | Secret key not stored in plain text |           |
| Update access key          | Change key, apply | Credential updated                  |           |
| Delete credential          | Destroy           | Credential removed                  |           |

---

#### 4.3 seqera_azure_credential

| Test Case               | Steps                        | Expected Result    | Pass/Fail |
| ----------------------- | ---------------------------- | ------------------ | --------- |
| Create Azure credential | Apply with tenant/client IDs | Credential created |           |
| Update credential       | Change values, apply         | Credential updated |           |
| Delete credential       | Destroy                      | Credential removed |           |

---

#### 4.4 seqera_google_credential

| Test Case             | Steps                           | Expected Result    | Pass/Fail |
| --------------------- | ------------------------------- | ------------------ | --------- |
| Create GCP credential | Apply with service account JSON | Credential created |           |
| Update credential     | Change values, apply            | Credential updated |           |
| Delete credential     | Destroy                         | Credential removed |           |

---

#### 4.5 seqera_github_credential

| Test Case                | Steps               | Expected Result    | Pass/Fail |
| ------------------------ | ------------------- | ------------------ | --------- |
| Create GitHub credential | Apply with PAT      | Credential created |           |
| Update token             | Change token, apply | Credential updated |           |
| Delete credential        | Destroy             | Credential removed |           |

---

#### 4.6 seqera_gitlab_credential

| Test Case                | Steps               | Expected Result    | Pass/Fail |
| ------------------------ | ------------------- | ------------------ | --------- |
| Create GitLab credential | Apply with token    | Credential created |           |
| Update token             | Change token, apply | Credential updated |           |
| Delete credential        | Destroy             | Credential removed |           |

---

#### 4.7 seqera_bitbucket_credential

| Test Case                   | Steps                   | Expected Result    | Pass/Fail |
| --------------------------- | ----------------------- | ------------------ | --------- |
| Create Bitbucket credential | Apply with app password | Credential created |           |
| Update credential           | Change values, apply    | Credential updated |           |
| Delete credential           | Destroy                 | Credential removed |           |

---

#### 4.8 seqera_gitea_credential

| Test Case               | Steps        | Expected Result    | Pass/Fail |
| ----------------------- | ------------ | ------------------ | --------- |
| Create Gitea credential | Apply config | Credential created |           |
| Delete credential       | Destroy      | Credential removed |           |

---

#### 4.9 seqera_codecommit_credential

| Test Case                    | Steps        | Expected Result    | Pass/Fail |
| ---------------------------- | ------------ | ------------------ | --------- |
| Create CodeCommit credential | Apply config | Credential created |           |
| Delete credential            | Destroy      | Credential removed |           |

---

#### 4.10 seqera_ssh_credential

| Test Case              | Steps                  | Expected Result               | Pass/Fail |
| ---------------------- | ---------------------- | ----------------------------- | --------- |
| Create SSH credential  | Apply with private key | Credential created            |           |
| Verify key not exposed | Check state            | Private key not in plain text |           |
| Delete credential      | Destroy                | Credential removed            |           |

---

#### 4.11 seqera_kubernetes_credential

| Test Case             | Steps                 | Expected Result    | Pass/Fail |
| --------------------- | --------------------- | ------------------ | --------- |
| Create K8s credential | Apply with kubeconfig | Credential created |           |
| Delete credential     | Destroy               | Credential removed |           |

---

#### 4.12 seqera_container_registry_credential

| Test Case                  | Steps                         | Expected Result    | Pass/Fail |
| -------------------------- | ----------------------------- | ------------------ | --------- |
| Create registry credential | Apply with registry URL/creds | Credential created |           |
| Delete credential          | Destroy                       | Credential removed |           |

---

#### 4.13 seqera_tower_agent_credential

| Test Case                     | Steps        | Expected Result    | Pass/Fail |
| ----------------------------- | ------------ | ------------------ | --------- |
| Create Tower Agent credential | Apply config | Credential created |           |
| Delete credential             | Destroy      | Credential removed |           |

---

### 5. Compute Environment Resources

#### 5.1 seqera_compute_env (Generic)

| Test Case          | Steps                        | Expected Result     | Pass/Fail |
| ------------------ | ---------------------------- | ------------------- | --------- |
| Create compute env | Apply config                 | Compute env created |           |
| Read compute env   | Plan after create            | No changes          |           |
| Update compute env | Change mutable fields, apply | Compute env updated |           |
| Delete compute env | Destroy                      | Compute env removed |           |

---

#### 5.2 seqera_aws_compute_env

| Test Case              | Steps                 | Expected Result                       | Pass/Fail |
| ---------------------- | --------------------- | ------------------------------------- | --------- |
| Create AWS compute env | Apply with AWS config | Compute env created                   |           |
| Verify in AWS console  | Check Batch/ECS       | Resources created                     |           |
| Delete compute env     | Destroy               | Compute env and AWS resources removed |           |

---

#### 5.3 seqera_aws_batch_ce

| Test Case       | Steps                   | Expected Result | Pass/Fail |
| --------------- | ----------------------- | --------------- | --------- |
| Create Batch CE | Apply with Batch config | CE created      |           |
| Delete CE       | Destroy                 | CE removed      |           |

---

#### 5.4 seqera_primary_compute_env

| Test Case         | Steps                   | Expected Result      | Pass/Fail |
| ----------------- | ----------------------- | -------------------- | --------- |
| Set primary CE    | Apply with CE reference | CE marked as primary |           |
| Change primary CE | Apply with different CE | Primary updated      |           |

---

### 6. Pipeline Resources

#### 6.1 seqera_pipeline

**Test Configuration:**

```hcl
resource "seqera_pipeline" "test" {
  workspace_id = seqera_workspace.test.id
  name         = "qa-test-pipeline"
  description  = "Pipeline for QA testing"
  repository   = "https://github.com/nextflow-io/hello"
}
```

| Test Case          | Steps                                                 | Expected Result     | Pass/Fail |
| ------------------ | ----------------------------------------------------- | ------------------- | --------- |
| Create pipeline    | Apply config                                          | Pipeline created    |           |
| Read pipeline      | Plan after create                                     | No changes          |           |
| Update description | Change description, apply                             | Description updated |           |
| Update repository  | Change repo, apply                                    | Repository updated  |           |
| Delete pipeline    | Destroy                                               | Pipeline removed    |           |
| Import pipeline    | `terraform import seqera_pipeline.test <pipeline_id>` | State imported      |           |

---

#### 6.2 seqera_pipeline_secret

**Test Configuration:**

```hcl
resource "seqera_pipeline_secret" "test" {
  workspace_id = seqera_workspace.test.id
  name         = "QA_TEST_SECRET"
  value        = "secret-value-123"
}
```

| Test Case                 | Steps               | Expected Result  | Pass/Fail |
| ------------------------- | ------------------- | ---------------- | --------- |
| Create secret             | Apply config        | Secret created   |           |
| Verify value not in state | Check state file    | Value not stored |           |
| Update value              | Change value, apply | Secret updated   |           |
| Delete secret             | Destroy             | Secret removed   |           |

---

### 7. Dataset Resources

#### 7.1 seqera_datasets

**Test Configuration:**

```hcl
resource "seqera_datasets" "test" {
  workspace_id = seqera_workspace.test.id
  name         = "qa-test-dataset"
  description  = "Dataset for QA testing"
}
```

| Test Case          | Steps                     | Expected Result     | Pass/Fail |
| ------------------ | ------------------------- | ------------------- | --------- |
| Create dataset     | Apply config              | Dataset created     |           |
| Read dataset       | Plan after create         | No changes          |           |
| Update description | Change description, apply | Description updated |           |
| Delete dataset     | Destroy                   | Dataset removed     |           |

---

#### 7.2 seqera_dataset_version

**Prerequisites:** Create a test CSV file: `echo "col1,col2\nval1,val2" > test.csv`

**Test Configuration:**

```hcl
resource "seqera_dataset_version" "test" {
  workspace_id = seqera_workspace.test.id
  dataset_id   = seqera_datasets.test.id
  file_path    = "test.csv"
  has_header   = true
}
```

| Test Case          | Steps                                                                                | Expected Result                 | Pass/Fail |
| ------------------ | ------------------------------------------------------------------------------------ | ------------------------------- | --------- |
| Create version     | Apply config                                                                         | Version 1 uploaded              |           |
| Read version       | Plan after create                                                                    | No changes                      |           |
| File unchanged     | Plan again                                                                           | No changes                      |           |
| File changed       | Modify file, plan                                                                    | Replacement planned             |           |
| File changed apply | Modify file, apply                                                                   | New version created             |           |
| Delete version     | Destroy                                                                              | Version disabled                |           |
| Import version     | `terraform import seqera_dataset_version.test <workspace_id>/<dataset_id>/<version>` | State imported (file_path null) |           |
| has_header=false   | Apply with has_header=false                                                          | Version created without header  |           |

---

### 8. Other Resources

#### 8.1 seqera_action

| Test Case     | Steps                | Expected Result | Pass/Fail |
| ------------- | -------------------- | --------------- | --------- |
| Create action | Apply config         | Action created  |           |
| Update action | Change config, apply | Action updated  |           |
| Delete action | Destroy              | Action removed  |           |

---

#### 8.2 seqera_labels

| Test Case    | Steps                    | Expected Result | Pass/Fail |
| ------------ | ------------------------ | --------------- | --------- |
| Create label | Apply config             | Label created   |           |
| Update label | Change name/value, apply | Label updated   |           |
| Delete label | Destroy                  | Label removed   |           |

---

#### 8.3 seqera_data_link

| Test Case        | Steps                | Expected Result   | Pass/Fail |
| ---------------- | -------------------- | ----------------- | --------- |
| Create data link | Apply config         | Data link created |           |
| Update data link | Change config, apply | Data link updated |           |
| Delete data link | Destroy              | Data link removed |           |

---

#### 8.4 seqera_studios

| Test Case     | Steps                | Expected Result | Pass/Fail |
| ------------- | -------------------- | --------------- | --------- |
| Create studio | Apply config         | Studio created  |           |
| Update studio | Change config, apply | Studio updated  |           |
| Delete studio | Destroy              | Studio removed  |           |

---

#### 8.5 seqera_tokens

| Test Case          | Steps         | Expected Result        | Pass/Fail |
| ------------------ | ------------- | ---------------------- | --------- |
| Create token       | Apply config  | Token created          |           |
| Verify token value | Check outputs | Token value accessible |           |
| Delete token       | Destroy       | Token revoked          |           |

---

#### 8.6 seqera_workflows

| Test Case              | Steps             | Expected Result         | Pass/Fail |
| ---------------------- | ----------------- | ----------------------- | --------- |
| Create/launch workflow | Apply config      | Workflow launched       |           |
| Read workflow          | Plan after create | No changes              |           |
| Delete workflow        | Destroy           | Workflow record removed |           |

---

## Data Sources Testing

### 1. data.seqera_organization_member

**Test Configuration:**

```hcl
data "seqera_organization_member" "test" {
  org_id = seqera_orgs.test.id
  email  = "existing-user@example.com"
}

output "member_id" {
  value = data.seqera_organization_member.test.member_id
}
```

| Test Case                  | Steps                    | Expected Result         | Pass/Fail |
| -------------------------- | ------------------------ | ----------------------- | --------- |
| Lookup existing member     | Apply config             | Returns member details  |           |
| Lookup non-existent member | Apply with invalid email | Error: Member Not Found |           |
| Verify all attributes      | Check outputs            | All fields populated    |           |

---

### 2. data.seqera_workspace

**Test Configuration:**

```hcl
data "seqera_workspace" "test" {
  org_id = seqera_orgs.test.id
  name   = "existing-workspace"
}

output "workspace_id" {
  value = data.seqera_workspace.test.workspace_id
}
```

| Test Case                     | Steps                   | Expected Result            | Pass/Fail |
| ----------------------------- | ----------------------- | -------------------------- | --------- |
| Lookup existing workspace     | Apply config            | Returns workspace details  |           |
| Lookup non-existent workspace | Apply with invalid name | Error: Workspace Not Found |           |
| Verify all attributes         | Check outputs           | All fields populated       |           |

---

### 3. data.seqera_workspace_participant

**Test Configuration:**

```hcl
data "seqera_workspace_participant" "test" {
  org_id       = seqera_orgs.test.id
  workspace_id = data.seqera_workspace.test.workspace_id
  email        = "existing-participant@example.com"
}
```

| Test Case                       | Steps                    | Expected Result              | Pass/Fail |
| ------------------------------- | ------------------------ | ---------------------------- | --------- |
| Lookup existing participant     | Apply config             | Returns participant details  |           |
| Lookup non-existent participant | Apply with invalid email | Error: Participant Not Found |           |
| Verify role returned            | Check outputs            | Role field populated         |           |

---

### 4. data.seqera_pipeline

**Test Configuration:**

```hcl
data "seqera_pipeline" "test" {
  workspace_id = seqera_workspace.test.id
  name         = "existing-pipeline"
}
```

| Test Case                    | Steps                   | Expected Result            | Pass/Fail |
| ---------------------------- | ----------------------- | -------------------------- | --------- |
| Lookup existing pipeline     | Apply config            | Returns pipeline details   |           |
| Lookup non-existent pipeline | Apply with invalid name | Error: Pipeline Not Found  |           |
| Verify repository returned   | Check outputs           | Repository field populated |           |

---

### 5. data.seqera_pipeline_secret

**Test Configuration:**

```hcl
data "seqera_pipeline_secret" "test" {
  workspace_id = seqera_workspace.test.id
  name         = "EXISTING_SECRET"
}
```

| Test Case                  | Steps                   | Expected Result         | Pass/Fail |
| -------------------------- | ----------------------- | ----------------------- | --------- |
| Lookup existing secret     | Apply config            | Returns secret metadata |           |
| Verify value NOT returned  | Check outputs           | Value field not present |           |
| Lookup non-existent secret | Apply with invalid name | Error: Secret Not Found |           |

---

### 6. data.seqera_credentials

**Test Configuration:**

```hcl
data "seqera_credentials" "test" {
  workspace_id = seqera_workspace.test.id
}
```

| Test Case                         | Steps         | Expected Result             | Pass/Fail |
| --------------------------------- | ------------- | --------------------------- | --------- |
| List credentials                  | Apply config  | Returns list of credentials |           |
| Verify sensitive data not exposed | Check outputs | Secrets not in output       |           |

---

### 7. data.seqera_data_links

**Test Configuration:**

```hcl
data "seqera_data_links" "test" {
  workspace_id = seqera_workspace.test.id
}
```

| Test Case       | Steps        | Expected Result            | Pass/Fail |
| --------------- | ------------ | -------------------------- | --------- |
| List data links | Apply config | Returns list of data links |           |

---

## Integration Testing Scenarios

### Scenario 1: Complete Organization Setup

Test creating a full organization hierarchy:

```hcl
# 1. Create organization
resource "seqera_orgs" "test" {
  name      = "integration-test-org"
  full_name = "Integration Test Organization"
}

# 2. Add organization member
resource "seqera_organization_member" "admin" {
  org_id = seqera_orgs.test.id
  email  = "admin@example.com"
  role   = "owner"
}

# 3. Create team
resource "seqera_teams" "dev" {
  org_id = seqera_orgs.test.id
  name   = "developers"
}

# 4. Add team member
resource "seqera_team_member" "dev_member" {
  org_id  = seqera_orgs.test.id
  team_id = seqera_teams.dev.id
  email   = "admin@example.com"
}

# 5. Create workspace
resource "seqera_workspace" "main" {
  org_id     = seqera_orgs.test.id
  name       = "main"
  visibility = "PRIVATE"
}

# 6. Add workspace participant
resource "seqera_workspace_participant" "admin" {
  org_id       = seqera_orgs.test.id
  workspace_id = seqera_workspace.main.id
  email        = "admin@example.com"
  role         = "admin"
}
```

| Test Case            | Steps             | Expected Result                | Pass/Fail |
| -------------------- | ----------------- | ------------------------------ | --------- |
| Create all resources | terraform apply   | All resources created in order |           |
| Verify dependencies  | Check Platform UI | Hierarchy correct              |           |
| Destroy in order     | terraform destroy | All resources removed          |           |

---

### Scenario 2: Pipeline with Credentials

```hcl
resource "seqera_github_credential" "repo" {
  workspace_id = var.workspace_id
  name         = "github-access"
  access_token = var.github_token
}

resource "seqera_pipeline" "nf_hello" {
  workspace_id  = var.workspace_id
  name          = "nf-hello"
  repository    = "https://github.com/nextflow-io/hello"
  credential_id = seqera_github_credential.repo.id
}
```

| Test Case                       | Steps             | Expected Result      | Pass/Fail |
| ------------------------------- | ----------------- | -------------------- | --------- |
| Create credential then pipeline | terraform apply   | Both created, linked |           |
| Pipeline uses credential        | Launch pipeline   | Successful git clone |           |
| Destroy pipeline first          | terraform destroy | Proper order         |           |

---

### Scenario 3: Dataset with Multiple Versions

```hcl
resource "seqera_datasets" "samples" {
  workspace_id = var.workspace_id
  name         = "sample-sheet"
}

resource "seqera_dataset_version" "v1" {
  workspace_id = var.workspace_id
  dataset_id   = seqera_datasets.samples.id
  file_path    = "samples_v1.csv"
}
```

| Test Case                  | Steps             | Expected Result     | Pass/Fail |
| -------------------------- | ----------------- | ------------------- | --------- |
| Create dataset and version | terraform apply   | Dataset with v1     |           |
| Update file content        | Modify CSV, apply | New version created |           |
| Destroy version            | terraform destroy | Version disabled    |           |

---

## Error Handling & Edge Cases

### Authentication Errors

| Test Case         | Steps             | Expected Result    | Pass/Fail |
| ----------------- | ----------------- | ------------------ | --------- |
| Invalid API token | Use wrong token   | Clear auth error   |           |
| Expired token     | Use expired token | Clear auth error   |           |
| Missing token     | Unset env var     | Clear config error |           |

### Permission Errors

| Test Case                          | Steps               | Expected Result | Pass/Fail |
| ---------------------------------- | ------------------- | --------------- | --------- |
| Create in org without access       | Apply as non-member | 403 error       |           |
| Modify resource without permission | Apply as viewer     | 403 error       |           |

### Resource Conflicts

| Test Case                   | Steps                   | Expected Result   | Pass/Fail |
| --------------------------- | ----------------------- | ----------------- | --------- |
| Duplicate organization name | Create same name twice  | Appropriate error |           |
| Duplicate member            | Add same user twice     | 409 conflict      |           |
| Duplicate workspace name    | Create same name in org | Appropriate error |           |

### External Changes (Drift)

| Test Case                    | Steps                    | Expected Result   | Pass/Fail |
| ---------------------------- | ------------------------ | ----------------- | --------- |
| Resource deleted externally  | Delete via UI, then plan | Shows recreation  |           |
| Resource modified externally | Modify via UI, then plan | Shows update/diff |           |

### Large Scale

| Test Case               | Steps                  | Expected Result       | Pass/Fail |
| ----------------------- | ---------------------- | --------------------- | --------- |
| Many org members (100+) | Create many members    | All managed correctly |           |
| Many workspaces (50+)   | Create many workspaces | All managed correctly |           |

---

## Test Results Template

### Test Run Information

- **Date:** YYYY-MM-DD
- **Tester:** Name
- **Provider Version:** x.x.x
- **Terraform Version:** x.x.x
- **Seqera Platform Version:** x.x.x

### Summary

| Category               | Total | Passed | Failed | Skipped |
| ---------------------- | ----- | ------ | ------ | ------- |
| Organization Resources |       |        |        |         |
| Workspace Resources    |       |        |        |         |
| Team Resources         |       |        |        |         |
| Credential Resources   |       |        |        |         |
| Compute Env Resources  |       |        |        |         |
| Pipeline Resources     |       |        |        |         |
| Dataset Resources      |       |        |        |         |
| Other Resources        |       |        |        |         |
| Data Sources           |       |        |        |         |
| Integration Scenarios  |       |        |        |         |
| Error Handling         |       |        |        |         |
| **TOTAL**              |       |        |        |         |

### Failed Tests

| Test Case | Resource | Error Message | Notes |
| --------- | -------- | ------------- | ----- |
|           |          |               |       |

### Notes

-
- ***

## Appendix: Quick Reference

### Import ID Formats

| Resource                     | Import Format                           | Example                        |
| ---------------------------- | --------------------------------------- | ------------------------------ |
| seqera_orgs                  | `<org_id>`                              | `12345`                        |
| seqera_organization_member   | `<org_id>/<email>`                      | `12345/user@example.com`       |
| seqera_workspace             | `<workspace_id>`                        | `67890`                        |
| seqera_workspace_participant | `<org_id>/<workspace_id>/<email>`       | `12345/67890/user@example.com` |
| seqera_teams                 | `<team_id>`                             | `11111`                        |
| seqera_team_member           | `<org_id>/<team_id>/<email>`            | `12345/11111/user@example.com` |
| seqera_dataset_version       | `<workspace_id>/<dataset_id>/<version>` | `67890/my-dataset/1`           |

### Role Values

| Resource                     | Valid Roles                                    |
| ---------------------------- | ---------------------------------------------- |
| seqera_organization_member   | `owner`, `member`, `collaborator`              |
| seqera_workspace_participant | `owner`, `admin`, `maintain`, `launch`, `view` |
