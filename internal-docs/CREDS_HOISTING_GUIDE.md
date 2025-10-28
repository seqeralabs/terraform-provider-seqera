# Credential Hoisting Guide for Seqera Terraform Provider

This guide explains how to properly implement credential hoisting in Speakeasy-generated Terraform providers, based on our working AWS credentials implementation and the [official Speakeasy hoisting guide](https://www.speakeasy.com/guides/terraform/hoisting).

## What is Hoisting?

Hoisting is a technique for "flattening" nested API structures to improve the Terraform user experience. By strategically placing the `x-speakeasy-entity` annotation, you can move nested credential fields to the root level of the Terraform resource schema while preserving the original API structure.

## Before and After

**Without Hoisting** (nested structure):
```hcl
resource "seqera_aws_credential" "example" {
  name = "my-creds"

  keys {
    access_key = "AKIA..."
    secret_key = "secret"
  }
}
```

**With Hoisting** (clean, flat structure):
```hcl
resource "seqera_aws_credential" "example" {
  name       = "my-creds"
  access_key = "AKIA..."
  secret_key = "secret"
}
```

## Implementation Pattern

Based on our working AWS credentials overlay (`overlays/credentials-aws.yaml`):

### 1. Schema Structure

```yaml
components:
  schemas:
    AWSCredential:
      required:
        - name
        - keys
      type: object
      properties:
        # Root-level metadata fields
        id:
          type: string
          x-speakeasy-name-override: credentialsId
          x-speakeasy-param-readonly: true
        name:
          type: string
          maxLength: 100
          x-speakeasy-param-force-new: true
        provider:
          type: string
          enum: [aws]
          default: aws
          x-speakeasy-name-override: providerType
          x-speakeasy-param-readonly: true

        # Metadata fields excluded from Terraform state
        deleted:
          type: boolean
          x-speakeasy-param-readonly: true
          x-speakeasy-terraform-ignore: true
        lastUsed:
          type: string
          format: date-time
          x-speakeasy-param-readonly: true
          x-speakeasy-terraform-ignore: true
        dateCreated:
          type: string
          format: date-time
          x-speakeasy-param-readonly: true
          x-speakeasy-terraform-ignore: true
        lastUpdated:
          type: string
          format: date-time
          x-speakeasy-param-readonly: true
          x-speakeasy-terraform-ignore: true

        # THE HOISTING HAPPENS HERE
        keys:
          x-speakeasy-entity: AWSCredential  # ← This causes hoisting!
          x-speakeasy-entity-description: |
            Manage AWS credentials in Seqera platform using this resource.

            AWS credentials store authentication information for accessing AWS services
            within the Seqera Platform workflows.
          type: object  # ← Must be inline, NOT $ref
          required:
            - accessKey
            - secretKey
          properties:
            accessKey:
              type: string
              minLength: 16
              maxLength: 128
              pattern: "^(AKIA|ASIA|AIDA)[A-Z0-9]{16,}$"
              description: AWS access key ID (required).
              x-speakeasy-param-force-new: true
              x-speakeasy-param-computed: false
              x-speakeasy-name-override: access_key
            secretKey:
              type: string
              minLength: 40
              description: AWS secret access key (required, sensitive).
              x-speakeasy-param-sensitive: true
              x-speakeasy-param-force-new: true
              x-speakeasy-param-computed: false
              x-speakeasy-name-override: secret_key
            assumeRoleArn:
              type: string
              pattern: "^arn:aws:iam::[0-9]{12}:role/.+$"
              description: IAM role ARN to assume (optional).
              x-speakeasy-param-force-new: true
              x-speakeasy-param-computed: false
              x-speakeasy-name-override: assume_role_arn
```

### 2. Request/Response Schemas

Keep these simple - only include what the API actually returns:

```yaml
CreateAWSCredentialsRequest:
  type: object
  properties:
    credentials:
      $ref: "#/components/schemas/AWSCredential"

CreateAWSCredentialsResponse:
  type: object
  properties:
    credentialsId:
      type: string

DescribeAWSCredentialsResponse:
  type: object
  properties:
    credentials:
      $ref: "#/components/schemas/AWSCredential"
```

**Important**: Don't include fields in create responses that are marked with `x-speakeasy-terraform-ignore: true`.

### 3. WriteOnly Fields via Overlay Actions

Use overlay actions to mark sensitive fields as `writeOnly`:

```yaml
actions:
  - target: $.components.schemas.AWSCredential.properties.keys.properties.secretKey
    update:
      writeOnly: true

  - target: $.components.schemas.AWSCredential.properties.keys.properties.assumeRoleArn
    update:
      writeOnly: true
```

This ensures these fields:
- Can be provided in create/update requests
- Will **never** be returned in API responses
- Are properly marked as sensitive in Terraform

## Key Rules for Hoisting

### ✅ DO:
1. Place `x-speakeasy-entity` on the nested object you want to hoist (e.g., `keys`)
2. Define properties **inline** using `type: object` and `properties:`
3. Use `x-speakeasy-name-override` for Terraform field names (e.g., `accessKey` → `access_key`)
4. Mark metadata fields with `x-speakeasy-terraform-ignore: true` if they shouldn't be in state
5. Use overlay actions to add `writeOnly: true` to sensitive fields
6. Keep create response schemas minimal - only what the API returns

### ❌ DON'T:
1. Don't use `$ref` for the nested object - it prevents hoisting
2. Don't place `x-speakeasy-entity` at the root schema level
3. Don't include `writeOnly` fields in response schemas
4. Don't include `x-speakeasy-terraform-ignore` fields in create responses
5. Don't add inline `writeOnly: true` - use overlay actions instead

## How It Works

When you place `x-speakeasy-entity` at the nested level:

1. **API Structure**: Preserved as-is with nested `keys` object
   ```json
   {
     "name": "my-creds",
     "keys": {
       "accessKey": "AKIA...",
       "secretKey": "secret"
     }
   }
   ```

2. **Terraform Schema**: Fields are hoisted to root level
   ```go
   type AWSCredentialResourceModel struct {
       Name       types.String `tfsdk:"name"`
       AccessKey  types.String `tfsdk:"access_key"`
       SecretKey  types.String `tfsdk:"secret_key"`
   }
   ```

3. **Speakeasy**: Automatically handles marshaling/unmarshaling between the two structures

## Common Patterns

### Read-only Computed Fields
```yaml
dateCreated:
  type: string
  format: date-time
  x-speakeasy-param-readonly: true
  x-speakeasy-terraform-ignore: true
```

### Sensitive Credentials
```yaml
# In schema definition:
secretKey:
  type: string
  x-speakeasy-param-sensitive: true

# In overlay actions:
- target: $.components.schemas.AWSCredential.properties.keys.properties.secretKey
  update:
    writeOnly: true
```

### Immutable Properties
```yaml
name:
  type: string
  x-speakeasy-param-force-new: true
```

## Testing

After implementing hoisting:

1. **Regenerate**: `speakeasy run --skip-versioning`
2. **Verify Schema**: Check that fields appear at root level in generated resource file
3. **Test Plan**: Run `terraform plan` in test directory
4. **Check State**: Verify fields are correctly populated after apply

## Reference Implementation

See `overlays/credentials-aws.yaml` (lines 190-337) for the complete, working implementation of credential hoisting in this provider.

## Resources

- [Speakeasy Hoisting Guide](https://www.speakeasy.com/guides/terraform/hoisting)
- [Speakeasy Extensions Reference](https://www.speakeasy.com/docs/speakeasy-reference/extensions)
- [Terraform Entity Mapping](https://www.speakeasy.com/docs/terraform/entity-mapping)
