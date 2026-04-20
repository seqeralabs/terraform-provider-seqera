# Speakeasy Extensions Reference

Complete reference of Speakeasy OpenAPI extensions for SDK and Terraform provider generation.

## General SDK Extensions

### Naming & Organization

#### x-speakeasy-name-override
- **Purpose**: Modify method, parameter, or class names in generated SDKs
- **Level**: Global or inline (method, parameter, class)
- **Use**: Rename identifiers while preserving API structure

#### x-speakeasy-group
- **Purpose**: Create custom namespaces for operations
- **Level**: Operation
- **Note**: Overrides default tag-based organization

#### x-speakeasy-ignore
- **Purpose**: Exclude specific methods from generated SDKs
- **Level**: Operation
- **Use**: Hide deprecated or internal endpoints

#### x-speakeasy-include
- **Purpose**: Force generation of orphaned schemas from components section
- **Level**: Schema
- **Use**: Include unused models for SDK users

### Enums

#### x-speakeasy-enums
- **Purpose**: Control generated enum members with alternative names
- **Level**: Schema/property
- **Format**: Map format recommended to prevent length mismatch errors

#### x-speakeasy-enum-descriptions
- **Purpose**: Attach descriptions to enum values for code documentation
- **Level**: Schema/property
- **Formats**: Array or map
- **Benefit**: IDE hints and documentation

#### x-speakeasy-enum-format
- **Purpose**: Control enum type (native enum vs union of strings)
- **Level**: Schema

### Documentation & Examples

#### x-speakeasy-usage-example
- **Purpose**: Feature specific methods in SDK README
- **Level**: Operation

#### x-speakeasy-example
- **Purpose**: Allow example values for request body properties
- **Level**: Property/schema
- **Note**: Overcomes OpenAPI spec limitation

#### x-speakeasy-docs
- **Purpose**: Configure language-specific comments in SDK
- **Level**: Operation/schema

### Runtime Behavior

#### x-speakeasy-retries
- **Purpose**: Enable retries globally or per-request with backoff strategies
- **Level**: Global or operation
- **Configuration**: Applies to specified HTTP status codes

#### x-speakeasy-pagination
- **Purpose**: Customize pagination (offset-based or cursor-based)
- **Level**: Operation

#### x-speakeasy-globals
- **Purpose**: Define SDK-level parameters populated across operations
- **Level**: Parameter/schema
- **Benefit**: Reduce method signature complexity

#### x-speakeasy-globals-hidden
- **Purpose**: Configure global parameters hidden from method signatures
- **Level**: Parameter

### Error Handling

#### x-speakeasy-errors
- **Purpose**: Override default error-handling behavior
- **Level**: Response, status code, or schema

#### x-speakeasy-error-message
- **Purpose**: Designate specific response field containing primary error message
- **Level**: Schema property (within error responses)

### Server & Auth

#### x-speakeasy-server-id
- **Purpose**: Enable users to pick a server when instantiating SDK
- **Level**: Servers array

#### x-speakeasy-overridable-scopes
- **Purpose**: Permit runtime OAuth scope override for authorization code flow
- **Level**: Security scheme
- **Requirement**: Adds optional scope field to security model

#### x-speakeasy-token-endpoint-additional-properties
- **Purpose**: Define custom fields for OAuth token endpoint requests
- **Level**: Security scheme
- **Use**: Support non-standard OAuth implementations

### Deprecation

#### x-speakeasy-deprecation-message
- **Purpose**: Add contextual deprecation messaging
- **Level**: Operation, parameter, or schema

#### x-speakeasy-deprecation-replacement
- **Purpose**: Specify recommended replacement operation
- **Level**: Operation

### Advanced

#### x-speakeasy-type-override
- **Purpose**: Force schema to be treated as arbitrary data type
- **Level**: Schema
- **Use**: Accept unstructured or dynamic JSON objects

#### x-speakeasy-max-method-params
- **Purpose**: Set maximum parameter count before converting to request object
- **Level**: Operation
- **Benefit**: Manage method signature complexity

#### x-speakeasy-param-encoding-override
- **Purpose**: Path parameters appear in URL with reserved characters unencoded
- **Level**: Parameter
- **Value**: Set to `true` to disable encoding
- **Use**: APIs requiring literal reserved characters in URLs

#### x-speakeasy-mcp
- **Purpose**: Customize how API operations are exposed as MCP tools
- **Level**: Operation
- **Properties**: disabled, name, title, scopes, description, destructiveHint, idempotentHint, openWorldHint, readOnlyHint

#### x-speakeasy-extension-rewrite
- **Purpose**: Map vendor-specific extensions to Speakeasy extensions
- **Benefit**: Reuse existing OpenAPI specs without modification

---

## Terraform-Specific Extensions

### Resource Mapping

#### x-speakeasy-entity
- **Purpose**: Map API entities to Terraform resources
- **Level**: Schema object
- **Use**: Annotate objects to create Terraform resources

#### x-speakeasy-entity-operation
- **Purpose**: Associate CRUD operations with Terraform resource lifecycle
- **Level**: Operation
- **Values**: create, read, update, delete

#### x-speakeasy-entity-version
- **Purpose**: Specify Terraform resource schema version for state migration
- **Level**: Schema
- **Note**: Use sparingly; adding/removing attributes doesn't require versioning

#### x-speakeasy-entity-description
- **Purpose**: Provide description for the Terraform resource
- **Level**: Schema object
- **Use**: Documentation shown in Terraform provider docs

### Property Constraints

#### x-speakeasy-param-force-new
- **Purpose**: Trigger resource recreation when property value changes
- **Level**: Property
- **Use**: Immutable properties

#### x-speakeasy-param-computed
- **Purpose**: Mark properties as computed (allow unknown values after apply)
- **Level**: Property
- **Caveat**: API must not modify computed values vs configuration

#### x-speakeasy-param-optional
- **Purpose**: Force property to be optional, overriding JSON Schema requirements
- **Level**: Property

#### x-speakeasy-param-readonly
- **Purpose**: Mark properties as read-only, preventing user modifications
- **Level**: Property
- **Use**: API-managed fields like IDs, timestamps

#### x-speakeasy-param-sensitive
- **Purpose**: Hide sensitive properties from Terraform console output
- **Level**: Property
- **Use**: Passwords, API keys, secrets

#### x-speakeasy-param-suppress-computed-diff
- **Purpose**: Indicate property never changes after creation
- **Level**: Property
- **Benefit**: Reduce unknown value output in plans

### Validation & Logic

#### x-speakeasy-plan-modifiers
- **Purpose**: Add custom logic to Terraform plan operations
- **Level**: Property
- **Use**: Defaults or replacement decisions

#### x-speakeasy-plan-validators
- **Purpose**: Enforce custom validation during planning phase
- **Level**: Property

#### x-speakeasy-conflicts-with
- **Purpose**: Prevent incompatible property combinations
- **Level**: Property

#### x-speakeasy-required-with
- **Purpose**: Indicate mutually necessary properties
- **Level**: Property

#### x-speakeasy-xor-with
- **Purpose**: Designate mutually exclusive property groups
- **Level**: Property

### State Management

#### x-speakeasy-soft-delete-property
- **Purpose**: Auto-remove and recreate resource when property is not null
- **Level**: Property
- **Use**: Detect soft-delete markers

#### x-speakeasy-terraform-ignore
- **Purpose**: Exclude properties from Terraform state management
- **Level**: Property

#### x-speakeasy-terraform-plan-only
- **Purpose**: Use only plan values during updates, ignoring prior state
- **Level**: Property

### Data Mapping

#### x-speakeasy-terraform-alias-to
- **Purpose**: Remap API response data to different property names
- **Level**: Property
- **Use**: Transform API field names to Terraform property names

#### x-speakeasy-match
- **Purpose**: Align API parameter names with Terraform state properties
- **Level**: Parameter

#### x-speakeasy-wrapped-attribute
- **Purpose**: Wrap API response data in Terraform schemas (primarily for arrays or additional operation data)
- **Level**: Property/schema
- **Use**: Control wrapper attribute name for array responses or multiple operation data
- **Note**: ⚠️ NOT for flattening nested credential structures. Use entity annotation placement instead.

### Custom Types

#### x-speakeasy-terraform-custom-type
- **Purpose**: Substitute terraform-plugin-framework custom types for base types
- **Level**: Property

---

## Important Notes

### **x-speakeasy-param-path DOES NOT EXIST**
This extension is NOT documented in Speakeasy. Do not use it.

### Terraform Property Mapping
For Terraform providers, use these for field mapping:
- `x-speakeasy-terraform-alias-to` - Remap response data
- `x-speakeasy-match` - Align parameter names
- `x-speakeasy-name-override` - Rename at schema level

### Hoisting Nested Structures in Terraform
To hoist nested API structures to the root level while preserving API compatibility:
1. Place `x-speakeasy-entity` annotation at the nested property level (e.g., `keys` property)
2. **CRITICAL**: Define properties **inline** using `type: object` and `properties:` (NOT using `$ref`)
3. Use `x-speakeasy-name-override` for Terraform field names

**Why this works:**
- Terraform schema gets hoisted fields at root level
- API structure preserved with nested objects in JSON
- Speakeasy handles marshaling/unmarshaling automatically

See `internal-docs/CREDS_HOISTING_GUIDE.md` for complete examples.

### Common Patterns

**Read-only computed fields:**
```yaml
dateCreated:
  type: string
  x-speakeasy-param-readonly: true
  x-speakeasy-param-suppress-computed-diff: true
```

**Sensitive credentials:**
```yaml
secretKey:
  type: string
  x-speakeasy-param-sensitive: true
```

**Immutable properties:**
```yaml
region:
  type: string
  x-speakeasy-param-force-new: true
```

---

## References

- [Speakeasy Extensions Documentation](https://www.speakeasy.com/docs/speakeasy-reference/extensions)
- [Terraform Provider Generation](https://www.speakeasy.com/docs/terraform)
- [Entity Mapping](https://www.speakeasy.com/docs/terraform/entity-mapping)
