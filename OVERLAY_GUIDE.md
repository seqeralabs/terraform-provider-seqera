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
- `examples/resources/seqera_labels/resource.tf` - Comprehensive label examples
- `examples/resources/seqera_orgs/resource.tf` - Simple organization examples
- `internal/validators/boolvalidators/label_is_default_validator.go` - Boolean validator
- `internal/validators/stringvalidators/label_value_resource_validator.go` - String validator with composition
