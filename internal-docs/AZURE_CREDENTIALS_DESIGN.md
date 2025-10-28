# Azure Credentials Design - Multi-Mode Support

## Overview

Azure credentials in Seqera Platform support three authentication modes:
1. **Shared Key** - Direct account key authentication (current implementation)
2. **Entra** - Azure Entra (formerly Azure AD) service principal authentication
3. **Cloud** - Azure Entra with Cloud-specific configurations

## Schema Design

### Fields Structure

```yaml
AzureCredential:
  properties:
    # Common fields (all modes)
    id: credentialsId
    name: string (required)
    provider: "azure" (computed)

    keys:
      mode: string (required) # "shared" | "entra" | "cloud"

      # Common to all modes
      batchName: string (required)
      storageName: string (required)  # "Blob Storage account name"

      # Shared key mode only
      batchKey: string (optional, sensitive, writeOnly)
      storageKey: string (optional, sensitive, writeOnly)

      # Entra/Cloud mode only
      tenantId: string (conditional required, writeOnly)
      clientId: string (conditional required, writeOnly)
      clientSecret: string (conditional required, sensitive, writeOnly)
```

## Validation Rules

### Rule 1: Mode-Based Required Fields

**When mode = "shared":**
- `batchName`: required
- `storageName`: required
- `batchKey`: optional
- `storageKey`: optional
- `tenantId`: must NOT be set
- `clientId`: must NOT be set
- `clientSecret`: must NOT be set

**When mode = "entra" or "cloud":**
- `batchName`: required
- `storageName`: required
- `tenantId`: required
- `clientId`: required
- `clientSecret`: required
- `batchKey`: must NOT be set
- `storageKey`: must NOT be set

### Rule 2: At Least One Authentication Method

For "shared" mode:
- At least one of `batchKey` or `storageKey` should be provided (warning, not error)

For "entra"/"cloud" mode:
- All three fields (`tenantId`, `clientId`, `clientSecret`) must be provided

## Implementation Plan

### 1. Create Custom Validators

Create validators in `internal/validators/stringvalidators/`:

- `azure_credential_mode_validator.go` - Validates mode field
- `azure_credential_shared_key_validator.go` - Validates shared key fields
- `azure_credential_entra_validator.go` - Validates Entra fields

### 2. Update Azure Overlay

Update `overlays/credentials-azure.yaml` to include:
- `mode` field as a discriminator
- All Entra-specific fields
- Proper `writeOnly` on all sensitive fields

### 3. Terraform User Experience

**Shared Key Mode:**
```hcl
resource "seqera_azure_credential" "shared" {
  name         = "azure-shared"
  workspace_id = seqera_workspace.main.id

  mode         = "shared"
  batch_name   = "myazurebatch"
  storage_name = "myazurestorage"
  batch_key    = var.azure_batch_key
  storage_key  = var.azure_storage_key
}
```

**Entra Mode:**
```hcl
resource "seqera_azure_credential" "entra" {
  name         = "azure-entra"
  workspace_id = seqera_workspace.main.id

  mode          = "entra"
  batch_name    = "myazurebatch"
  storage_name  = "myazurestorage"
  tenant_id     = var.azure_tenant_id
  client_id     = var.azure_client_id
  client_secret = var.azure_client_secret
}
```

## Open Questions

1. **Default Mode**: Should we default to "shared" for backwards compatibility?
2. **API Field Name**: Is the discriminator field called `mode`, `type`, `authMode`, or something else?
3. **Cloud vs Entra**: Are there any differences between "cloud" and "entra" modes beyond the name?

## Next Steps

1. Verify actual API field names by testing or checking latest API documentation
2. Implement custom validators following the labels pattern
3. Update Azure overlay with all fields and validators
4. Regenerate and test both modes
