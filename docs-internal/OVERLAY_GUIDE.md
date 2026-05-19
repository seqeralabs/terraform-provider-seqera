# Overlay and Resource Documentation Guide

This guide documents best practices for creating and maintaining Speakeasy overlay files and Terraform resource examples in this provider.

## Table of Contents
- [Overlay File Structure](#overlay-file-structure)
- [Field Management](#field-management)
- [Resource Examples](#resource-examples)
- [Custom Validators](#custom-validators)
- [Documentation](#documentation)

## Overlay File Structure

Overlay files should follow a consistent structure with clear sections:

```yaml
overlay: 1.0.0
x-speakeasy-jsonpath: rfc9535
info:
  title: [Resource Name] Overlay
  version: 0.0.0

# ==============================================================================
# [RESOURCE NAME] RESOURCE OVERLAY
# ==============================================================================
#
# Brief description of the resource and its purpose
#
# TERRAFORM EXAMPLES:
# -------------------
#
# Example 1: Basic usage
# resource "seqera_resource" "basic" {
#   required_field = "value"
# }
#
# Example 2: With optional fields
# resource "seqera_resource" "complete" {
#   required_field = "value"
#   optional_field = "value"
# }
#
# ==============================================================================

actions:
  # ============================================================================
  # ENTITY CONFIGURATION
  # ============================================================================

  # Entity schema configuration
  - target: $["components"]["schemas"]["ResourceSchema"]
    update:
      x-speakeasy-entity: ResourceName
      x-speakeasy-entity-description: |
        Description for the resource documentation

  # ============================================================================
  # ENTITY OPERATIONS (CRUD)
  # ============================================================================

  # CREATE - POST /resource
  - target: $["paths"]["/resource"]["post"]
    update:
      x-speakeasy-entity-operation: ResourceName#create

  # READ - GET /resource/{id}
  - target: $["paths"]["/resource/{id}"]["get"]
    update:
      x-speakeasy-entity-operation: ResourceName#read

  # UPDATE - PUT /resource/{id}
  - target: $["paths"]["/resource/{id}"]["put"]
    update:
      x-speakeasy-entity-operation: ResourceName#update

  # DELETE - DELETE /resource/{id}
  - target: $["paths"]["/resource/{id}"]["delete"]
    update:
      x-speakeasy-entity-operation: ResourceName#delete

  # ============================================================================
  # FIELD VALIDATORS
  # ============================================================================

  # Apply custom validators to fields
  - target: $.components.schemas.ResourceSchema.properties.fieldName
    update:
      description: "Field description with validation rules"
      x-speakeasy-plan-validators: CustomValidatorName

  # ============================================================================
  # FIELD CONFIGURATION
  # ============================================================================

  # Optional/computed field configuration
  - target: $["components"]["schemas"]["ResourceSchema"]["properties"]["computedField"]
    update:
      x-speakeasy-param-optional: true

  # ============================================================================
  # SCHEMA DESCRIPTIONS
  # ============================================================================

  # Improve field descriptions for generated documentation
  - target: $.components.schemas.ResourceSchema.properties.fieldName
    update:
      description: Clear, detailed description of what this field does

  # ============================================================================
  # REQUEST EXAMPLES
  # ============================================================================

  # Provide realistic examples for documentation
  - target: $.components.schemas.ResourceSchema.properties.fieldName
    update:
      example: "example-value"

  # ============================================================================
  # CLEANUP - Remove unmanageable and internal fields
  # ============================================================================

  # Remove fields that cannot be managed via Terraform
  - target: $["components"]["schemas"]["ResourceSchema"]["properties"]["unmanagedField"]
    remove: true
```

## Field Management

### Fields to Remove

Remove fields that fall into these categories:

#### 1. Unmanageable Fields
Fields that require external actions (file uploads, special workflows):
```yaml
# Example: Logo fields requiring file upload via web UI
- target: $["components"]["schemas"]["Organization"]["properties"]["logoId"]
  remove: true

- target: $["components"]["schemas"]["OrganizationDbDto"]["properties"]["logoUrl"]
  remove: true
```

#### 2. Internal/System Fields
Fields managed by the platform that users should not set:
```yaml
# Example: Internal organization type and billing status
- target: $["components"]["schemas"]["OrganizationDbDto"]["properties"]["paying"]
  remove: true

- target: $["components"]["schemas"]["OrganizationDbDto"]["properties"]["type"]
  remove: true
```

#### 3. Deprecated Fields
Fields marked as deprecated in the API:
```yaml
# Example: Deprecated label fields
- target: $["components"]["schemas"]["LabelDbDto"]["properties"]["isDynamic"]
  remove: true
```

### Field Descriptions

Improve field descriptions to be clear and actionable:

**Good Examples:**
```yaml
- target: $.components.schemas.Organization.properties.name
  update:
    description: Short name or handle for the organization (used in URLs and resource paths). Required.

- target: $.components.schemas.Organization.properties.fullName
  update:
    description: Complete formal display name of the organization. Required.
```

**Bad Examples:**
```yaml
description: The name  # Too brief
description: Organization name field  # Redundant
description: Name of the organization in the system  # Vague
```

## Response Field Mapping

### Adding a `.id` Alias to a Resource (Additive — Preferred)

Most Seqera entities expose their primary key under a name like `pipelineId`,
`credentialsId`, `labelId`, `computeEnvId`. Terraform users expect to be able
to write `seqera_pipeline.foo.id` regardless. The right tool for this is
`x-speakeasy-transform-from-api` — it adds a synthetic `id` field to the
response *alongside* the existing `{entity}Id`, with no wire-format change
and no state migration.

**Skip resources that lack a single primary key.** Not every entity has one.
`seqera_custom_role` is addressed server-side by the composite
`(org_id, name)` tuple — the user-facing identifier is the role NAME, and
that's what gets interpolated into `seqera_workspace_participant.role_names`.
Synthesizing an `id` (whether from the name or a composite encoding) would
imply a primary key that doesn't exist and confuse customers about how to
reference the role. `seqera_primary_compute_env` is similar — it's an
action-like resource that sets one CE as primary for a workspace, with no
entity identity of its own. Leave these resources without an `.id` alias.

#### The Additive Alias Pattern

```yaml
# Add a parallel `id` attribute that aliases `pipelineId`. The jq
# expression runs on response deserialization; the wire payload is
# unchanged. Existing state with `pipeline_id` keeps validating —
# `id` is purely additive.
- target: $.components.schemas.PipelineDbDto
  update:
    x-speakeasy-transform-from-api:
      jq: '. + { id: .pipelineId }'
    properties:
      id:
        type: integer
        format: int64
        description: Alias of `pipeline_id` for Terraform convention.
        x-speakeasy-param-readonly: true
```

The `properties.id` block under `update` declares the new attribute to
Speakeasy so the TF schema gains a Computed `id` field. The `jq`
expression populates it from `pipelineId` at read time.

**What you get:**
- `seqera_pipeline.foo.id` resolves (Int64) — TF convention satisfied
- `seqera_pipeline.foo.pipeline_id` still works — no breaking change
- Same value: `id == pipeline_id`
- Wire format unchanged (`pipelineId` over HTTP)
- No state upgrader needed — existing state is forward-compatible
- No downstream HCL or docs breakage

**If the update/create endpoints reject extra fields:** pair the transform
with an outbound strip so `id` isn't sent on requests:

```yaml
- target: $.components.schemas.UpdatePipelineRequest
  update:
    x-speakeasy-transform-to-api:
      jq: 'del(.id)'
```

In practice the Seqera API ignores unknown fields, so this is rarely
required — verify with `terraform apply` before adding it.

**Type alignment matters.** The synthetic `id` attribute's `type` and
`format` must match the source field. For string IDs (`credentialsId`,
`computeEnvId`, `workflowId`):

```yaml
properties:
  id:
    type: string
    description: Alias of `credentials_id` for Terraform convention.
    x-speakeasy-param-readonly: true
```

### The Rename Pattern (Destructive — Avoid for Existing Resources)

`x-speakeasy-name-override: id` paired with `x-speakeasy-match: id` on path
parameters does work, and is documented in the Speakeasy guide for *new*
resources. **Do not use it on a resource that already exists** — the rename
removes the original `{entity}_id` attribute, breaking:

- existing Terraform state (decode errors on the next plan — requires a
  custom `StateUpgrader` for every renamed resource)
- customer HCL referencing `seqera_*.foo.{entity}_id`
- generated docs / examples / showcase repo

Reach for the rename only when greenfielding a brand-new resource where
no consumer has the old shape. For everything else, use the additive
pattern above.

For reference, the rename pattern looks like:

```yaml
# DESTRUCTIVE — only use on greenfield resources.
- target: $.components.schemas.CreateResourceResponse.properties.credentialsId
  update:
    x-speakeasy-name-override: id

- target: $["paths"]["/resource/{credentialsId}"]["get"]["parameters"][?(@.name == "credentialsId")]
  update:
    x-speakeasy-match: id

- target: $["paths"]["/resource/{credentialsId}"]["put"]["parameters"][?(@.name == "credentialsId")]
  update:
    x-speakeasy-match: id

- target: $["paths"]["/resource/{credentialsId}"]["delete"]["parameters"][?(@.name == "credentialsId")]
  update:
    x-speakeasy-match: id
```

This rewrites the Go struct field to `ID` while keeping the JSON tag at
`credentialsId`, and threads the value into URL parameters via the match
extension.

#### Path Anchor Considerations

Ensure the path anchor in your overlay target matches the actual API path:

```yaml
# WRONG - Will fail if API uses different anchor
- target: $["paths"]["/credentials/{credentialsId}#gcp"]["get"]...

# RIGHT - Use the actual path anchor from the OpenAPI spec
- target: $["paths"]["/credentials/{credentialsId}#google"]["get"]...
```

**Common Path Anchors:**
- Google credentials: `#google` (not `#gcp`)
- Kubernetes credentials: `#k8s` (not `#kubernetes`)
- Tower Agent credentials: `#agent` (not `#tower-agent`)

## Hoisting Nested API Structures into Flat Terraform Resources

The Seqera API frequently returns credential-like payloads where the
meaningful fields live one level down, e.g.
`{ "name": "...", "provider": "aws", "keys": { "accessKey": "...",
"secretKey": "..." } }`. Terraform users expect those fields at the
resource root (`access_key`, `secret_key`) rather than under a `keys`
block. Two patterns exist for closing that gap.

### Pattern A: Transform-based hoisting (recommended for new work)

Use `x-speakeasy-transform-from-api` to flatten nested fields into the
top level on read, and `x-speakeasy-transform-to-api` to put them back
under the nested key on write. The wire format stays nested; the
Terraform schema stays flat.

```yaml
- target: $.components.schemas.AWSCredential
  update:
    x-speakeasy-entity: AWSCredential
    x-speakeasy-transform-from-api:
      jq: |
        . + {
          access_key:      .keys.accessKey,
          secret_key:      .keys.secretKey,
          assume_role_arn: .keys.assumeRoleArn,
          mode:            .keys.mode,
          external_id:     .keys.externalId
        } | del(.keys)
    x-speakeasy-transform-to-api:
      jq: |
        . + { keys: {
          accessKey:     .access_key,
          secretKey:     .secret_key,
          assumeRoleArn: .assume_role_arn,
          mode:          .mode,
          externalId:    .external_id
        }} | del(.access_key, .secret_key, .assume_role_arn,
                  .mode, .external_id)
    properties:
      access_key:
        type: string
        minLength: 16
        maxLength: 128
        pattern: ^(AKIA|ASIA|AIDA)[A-Z0-9]{16,}$
        example: AKIAIOSFODNN7EXAMPLE
        x-speakeasy-param-optional: true
        x-speakeasy-plan-validators: AWSCredentialKeysValidator
      secret_key:
        type: string
        minLength: 40
        x-speakeasy-param-sensitive: true
        x-speakeasy-terraform-write-only: true
        x-speakeasy-param-optional: true
        x-speakeasy-plan-validators: AWSCredentialKeysValidator
      # ... assume_role_arn, mode, external_id
```

**Why prefer this for new resources:**
- The `x-speakeasy-entity` annotation lives on the parent schema, where
  it belongs — no nested-entity gymnastics.
- Field naming happens once in the jq pipeline. No per-field
  `x-speakeasy-name-override`.
- All field metadata (validators, sensitivity, write-only, examples)
  sits at the entity root. Easy to scan, easy to extend.
- Adding a new hoisted field is a three-line change (one jq line each
  direction + the `properties` block).
- Wire format unchanged — backend never sees the flattened shape.

**Risks to manage:**
- **Bidirectional drift.** Forget to keep the `from-api` and `to-api`
  expressions in sync and writes silently corrupt the request body.
  Treat the two `jq` blocks as one unit; review them together.
- **Sensitive / write-only fields.** Annotations still go on the
  hoisted property in `properties:` — jq doesn't carry them across.
- **Field-count scaling.** jq grows linearly with the number of
  hoisted fields. Fine up to ~10; consider splitting into helper
  schemas beyond that.

### Pattern B: Nested-entity placement (legacy — not used anywhere today)

The older approach placed `x-speakeasy-entity` *inside* the nested
property (e.g. `AWSCredential.properties.keys`), defined the hoisted
fields inline under `keys.properties`, and used
`x-speakeasy-name-override` to snake_case each one.

Drawbacks compared to Pattern A:
- Entity annotation lives in an unusual location relative to the
  schema's natural root.
- Every nested field needs `x-speakeasy-name-override` for the snake_case
  TF attribute name.
- Schema structure cannot use `$ref` — properties must be inlined,
  which duplicates definitions if multiple resources share a payload
  shape.
- Adding or renaming a hoisted field touches both the nested location
  and any related path-anchor overlays.

All credentials + `seqera_managed_compute_ce` were migrated off Pattern B
in 0.40.0-RC6 (the generated TF schemas were byte-identical pre/post
migration). Pattern B is no longer in use anywhere in this provider.
Documented here for reference if you encounter it in older overlays or
historical PR review.

### User-Provided Optional Fields

When a field is user-provided and optional (not computed by the API), use `x-speakeasy-param-computed: false`:

```yaml
- target: $.components.schemas.ResourceSchema.properties.baseUrl
  update:
    type: string
    description: 'Optional base URL for self-hosted server'
    example: https://gitlab.mycompany.com
    x-speakeasy-name-override: base_url
    x-speakeasy-param-computed: false  # User provides this, API doesn't compute it
```

**Why This Matters:**
- `x-speakeasy-param-computed: true` tells Terraform to expect the API to return/compute the value
- If the API doesn't return it in the Create response, Terraform shows "unknown value" error
- `x-speakeasy-param-computed: false` makes it a simple optional field that stores user input

**When to Use:**
- Optional configuration fields (URLs, paths, flags)
- Fields where user input should be preserved as-is
- Fields not returned in Create responses

**When NOT to Use:**
- Fields that are actually computed by the API (timestamps, IDs, status)
- Fields marked as `x-speakeasy-param-readonly: true`

#### Complete Credential Resource Example

Here's a complete example showing all patterns together (from `overlays/credentials-github.yaml`):

```yaml
# Create operation with workspace parameter
- target: $.paths
  update:
    /credentials#github:
      post:
        x-speakeasy-entity-operation:
          terraform-resource: GithubCredential#create
        parameters:
        - name: workspaceId
          in: query
          schema:
            type: integer
            format: int64

# Schema definition
- target: $.components.schemas
  update:
    GithubCredential:
      properties:
        credentials_id:
          type: string
          description: Unique identifier for the credential (max 22 characters)
          x-speakeasy-param-readonly: true
        baseUrl:
          type: string
          description: 'Repository base URL for GitHub Enterprise (optional)'
          x-speakeasy-name-override: base_url
          x-speakeasy-param-computed: false  # User-provided optional field
        keys:
          type: object
          required:
          - username
          - accessToken
          properties:
            username:
              type: string
              description: GitHub username
            accessToken:
              type: string
              description: GitHub Personal Access Token
              x-speakeasy-param-sensitive: true

# Response schemas
- target: $.components.schemas
  update:
    CreateGithubCredentialsResponse:
      type: object
      properties:
        credentialsId:
          type: string


# Mark workspace parameter as force-new (requires replacement)
- target: $["paths"]["/credentials#github"]["post"]["parameters"][?(@.name == "workspaceId")]
  update:
    x-speakeasy-param-force-new: true
```

**Result:**
- Create returns `credentialsId`, properly mapped to Terraform's `id` field
- Read, Update, Delete operations use the `id` value for the path parameter
- `baseUrl` is an optional user-provided field, not computed by API
- Workspace changes require resource replacement

## Resource Examples

### File Location
Place custom examples in `examples/resources/seqera_[resource]/resource.tf`

### Protecting Custom Examples
Add custom examples to `.genignore` to prevent Speakeasy from overwriting:
```
# Custom examples that should be manually maintained
examples/resources/seqera_labels/resource.tf
examples/resources/seqera_orgs/resource.tf
```

### Example Structure

**Keep examples simple and focused on the resource itself:**

```terraform
# Resource Type Examples
#
# Brief description of what this resource does

# Example 1: Basic usage
# Minimal configuration with required fields only

resource "seqera_resource" "basic" {
  required_field = "value"
}

# Example 2: With optional fields
# Show additional configuration options

resource "seqera_resource" "complete" {
  required_field  = "value"
  optional_field1 = "value"
  optional_field2 = "value"
}
```

**Avoid:**
- Complex multi-resource examples in the primary resource documentation
- Cross-resource dependencies (save these for the dependent resource)
- Multiple variations showing every possible combination
- Emojis and ASCII art

**Example: Organizations should not show workspaces**
```terraform
# BAD - Don't do this in orgs examples
resource "seqera_orgs" "example" {
  name = "my-org"
}

resource "seqera_workspace" "example" {  # Belongs in workspace examples
  org_id = seqera_orgs.example.org_id
}

# GOOD - Keep it simple
resource "seqera_orgs" "example" {
  name        = "my-org"
  full_name   = "My Organization"
  description = "Organization for computational research"
}
```

**Show dependencies in the dependent resource:**
```terraform
# In examples/resources/seqera_workspace/resource.tf
resource "seqera_workspace" "example" {
  name      = "my-workspace"
  org_id    = seqera_orgs.parent.org_id  # Reference to parent
  full_name = "${seqera_orgs.parent.name}/my-workspace"
}
```

## Custom Validators

### When to Use Custom Validators

Use custom validators when:
1. Fields have conditional requirements based on other fields
2. Built-in Speakeasy validators (`x-speakeasy-conflicts-with`, `x-speakeasy-xor-with`) are insufficient
3. Complex validation logic is needed

### Validator Patterns

#### Boolean Validator Pattern
For validating boolean fields based on sibling fields:

```go
// File: internal/validators/boolvalidators/example_validator.go
type BoolExampleValidator struct{}

func (v BoolExampleValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
    // Skip if null, unknown, or false
    if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
        return
    }

    // Get sibling field
    var siblingValue types.Bool
    siblingPath := req.Path.ParentPath().AtName("sibling_field")
    resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &siblingValue)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Allow unknown values during plan phase
    if siblingValue.IsUnknown() {
        return
    }

    // Validation logic
    if siblingValue.IsNull() || !siblingValue.ValueBool() {
        resp.Diagnostics.AddAttributeError(
            req.Path,
            "Invalid Configuration",
            "field_name can only be true when sibling_field is true",
        )
    }
}
```

#### String Validator Pattern
For validating string fields with cross-field dependencies:

```go
// File: internal/validators/stringvalidators/example_validator.go
type StringExampleValidator struct{}

func (v StringExampleValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
    var siblingValue types.Bool
    siblingPath := req.Path.ParentPath().AtName("sibling_field")
    resp.Diagnostics.Append(req.Config.GetAttribute(ctx, siblingPath, &siblingValue)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Allow unknown values during plan phase (for_each, etc.)
    if siblingValue.IsUnknown() || req.ConfigValue.IsUnknown() {
        return
    }

    siblingIsTrue := !siblingValue.IsNull() && siblingValue.ValueBool()
    valueIsEmpty := req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == ""

    // Bidirectional validation
    if siblingIsTrue && valueIsEmpty {
        resp.Diagnostics.AddAttributeError(
            req.Path,
            "Missing Required Field",
            "field_name must be set when sibling_field is true",
        )
        return
    }

    if !valueIsEmpty && !siblingIsTrue {
        resp.Diagnostics.AddAttributeError(
            req.Path,
            "Invalid Configuration",
            "field_name can only be set when sibling_field is true",
        )
        return
    }

    // Apply format validation if value is present
    if !valueIsEmpty {
        formatValidator := FormatValidator()
        formatValidator.ValidateString(ctx, req, resp)
    }
}
```

#### Object Validator Pattern
For validating complex object structures with JSON content:

```go
// File: internal/validators/objectvalidators/example_validator.go
type ObjectExampleValidator struct{}

func (v ObjectExampleValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
    // Skip validation if object is null or unknown
    if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
        return
    }

    attrs := req.ConfigValue.Attributes()

    // Check if required field exists
    dataAttr, exists := attrs["data"]
    if !exists || dataAttr.IsNull() {
        resp.Diagnostics.AddAttributeError(
            req.Path.AtName("data"),
            "Missing Required Field",
            "The 'data' field is required and cannot be null",
        )
        return
    }

    // Get the string value
    stringValue, ok := dataAttr.(basetypes.StringValue)
    if !ok {
        resp.Diagnostics.AddAttributeError(
            req.Path.AtName("data"),
            "Invalid Data Type",
            "The 'data' field must be a string",
        )
        return
    }

    // Allow unknown values during plan phase
    if stringValue.IsUnknown() || stringValue.IsNull() {
        return
    }

    jsonData := stringValue.ValueString()
    if jsonData == "" {
        resp.Diagnostics.AddAttributeError(
            req.Path.AtName("data"),
            "Empty Value",
            "The 'data' field cannot be empty",
        )
        return
    }

    // Validate JSON structure
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
        resp.Diagnostics.AddAttributeError(
            req.Path.AtName("data"),
            "Invalid JSON",
            fmt.Sprintf("The 'data' field must contain valid JSON. Error: %s\n\nTip: Use file() to read from a JSON file", err.Error()),
        )
        return
    }

    // Validate required JSON fields
    requiredFields := []string{"field1", "field2"}
    var missingFields []string
    for _, field := range requiredFields {
        if _, exists := data[field]; !exists {
            missingFields = append(missingFields, field)
        }
    }

    if len(missingFields) > 0 {
        resp.Diagnostics.AddAttributeError(
            req.Path.AtName("data"),
            "Missing Required JSON Fields",
            fmt.Sprintf("Required fields: %v\nMissing fields: %v", requiredFields, missingFields),
        )
        return
    }
}
```

### Validator Best Practices

1. **Handle Unknown Values**: Always check for `IsUnknown()` to support Terraform's plan phase with `for_each`, `count`, etc.
2. **Early Returns**: Use early returns to simplify logic and avoid nested conditions
3. **Reuse Validators**: Have validators call other validators to avoid duplication
4. **Clear Error Messages**: Provide actionable error messages that explain the constraint
5. **Bidirectional Validation**: When fields depend on each other, validate both directions

### Applying Validators in Overlays

```yaml
- target: $.components.schemas.CreateRequest.properties.fieldName
  update:
    title: "Field Display Name"
    description: "Clear description including validation rules"
    example: "example-value"
    x-speakeasy-plan-validators: CustomValidatorName

# Apply to both Create and Update schemas
- target: $.components.schemas.UpdateRequest.properties.fieldName
  update:
    title: "Field Display Name"
    description: "Clear description including validation rules"
    example: "example-value"
    x-speakeasy-plan-validators: CustomValidatorName
```

## Documentation

### Generated Documentation Structure

After running `speakeasy run`, documentation is generated in `docs/resources/[resource].md` with:
1. Frontmatter (title, description)
2. Resource description
3. Example usage (from `examples/resources/seqera_[resource]/resource.tf`)
4. Schema (Required, Optional, Read-Only fields)
5. Import instructions

### Verifying Documentation

After regeneration, check:
1. Field descriptions are clear and helpful
2. Examples render correctly
3. Only manageable fields appear in the schema
4. Read-only fields are properly marked
5. Required vs optional fields are correct

### Common Issues

**Issue**: Nested schema showing in documentation
```yaml
# Solution: Remove from response schema
- target: $["components"]["schemas"]["ListResponse"]["properties"]["items"]
  remove: true
```

**Issue**: Field description missing
```yaml
# Solution: Add description in overlay
- target: $.components.schemas.Schema.properties.fieldName
  update:
    description: "Clear description of the field"
```

**Issue**: Internal fields exposed
```yaml
# Solution: Remove internal fields
- target: $["components"]["schemas"]["Schema"]["properties"]["internalField"]
  remove: true
```

## Workflow

### Making Changes

1. **Identify Fields to Manage**
   - Review the OpenAPI spec and identify unmanageable/internal fields
   - Check what fields actually work via the API

2. **Update Overlay**
   - Remove unmanageable fields
   - Add/improve field descriptions
   - Configure validators if needed
   - Organize into clear sections

3. **Create Custom Examples**
   - Write focused, simple examples in `examples/resources/seqera_[resource]/resource.tf`
   - Add to `.genignore` to protect from regeneration

4. **Regenerate**
   ```bash
   speakeasy run --skip-versioning
   ```

5. **Verify**
   - Check generated `docs/resources/[resource].md`
   - Build the provider: `go build -o terraform-provider-seqera`
   - Test with `terraform plan` in `examples/tests/`

6. **Commit**
   - Commit overlay changes
   - Commit custom examples
   - Commit `.genignore` updates
   - Commit generated code and docs

## Examples

See the following resources for complete examples:
- `overlays/labels.yaml` - Labels resource with custom validators
- `overlays/orgs.yaml` - Organizations resource with field cleanup
- `overlays/pipeline-secrets.yaml` - Pipeline secrets with sensitive fields
- `overlays/credentials-gcp.yaml` - Google credentials with object validator
- `examples/resources/seqera_labels/resource.tf` - Comprehensive label examples
- `examples/resources/seqera_orgs/resource.tf` - Simple organization examples
- `examples/resources/seqera_pipeline_secret/resource.tf` - Security-focused secret examples
- `examples/resources/seqera_google_credential/resource.tf` - GCP credential examples
- `internal/validators/boolvalidators/label_is_default_validator.go` - Boolean validator
- `internal/validators/stringvalidators/label_value_resource_validator.go` - String validator with composition
- `internal/validators/objectvalidators/google_keys_crdential_validator.go` - Object validator with JSON validation
