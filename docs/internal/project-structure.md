# Project Structure

This document describes the structure and organization of the Seqera Terraform Provider codebase.

## Overview

The Seqera Terraform Provider is an auto-generated codebase built using [Speakeasy](https://speakeasy.com) from the Seqera platform OpenAPI specifications. The provider enables management of Seqera Platform resources through Terraform configurations.

## Key Architecture Principles

- **Auto-generated codebase**: Generated using Speakeasy from OpenAPI specifications
- **Go-based Terraform provider**: Uses `terraform-plugin-framework` 
- **Immutable generated code**: Manual changes to internal files will be overwritten on next generation (Excluding custom validation logic)
- **Overlay-driven customization**: Provider behavior is customized through overlay files

## Directory Structure

### Core Generated Code

```
internal/
├── provider/           # Main provider implementation
│   ├── provider.go     # Core provider configuration and setup
│   ├── *_resource.go   # Resource implementations (seqera_*)
│   ├── *_data_source.go# Data source implementations
│   ├── *_sdk.go        # SDK integration layers
│   ├── types/          # Terraform schema type definitions
│   ├── reflect/        # Reflection utilities for type conversion
│   └── validators/     # Custom validation logic (These files can be manually edited)
└── sdk/                # Auto-generated SDK for Seqera API
    ├── *.go            # API client implementations
    ├── models/         # Model definitions
    │   ├── shared/     # Shared model types
    │   └── operations/ # Operation-specific types
    └── internal/       # SDK internal utilities
```

### Configuration and Schema Management

```
schemas/
├── seqera-api-latest.yml   # Base OpenAPI specification
├── seqera-final.yaml       # Modified OpenAPI spec with Speakeasy annotations
└── overlay.yaml            # Speakeasy overlay for customizations

.speakeasy/
├── gen.yaml               # Generation configuration used by speakeasy
├── workflow.yaml          # Workflow definition, this is what `speakeasy run` uses
└── out.openapi.yaml       # Processed OpenAPI specification after applying the overlay. 
```

### Documentation and Examples

```
docs/
├── index.md              # Main provider documentation
├── resources/            # Resource documentation
├── data-sources/         # Data source documentation
└── internal/             # Internal development documentation

examples/
├── tests/                # Test configurations for local development
└── provider/             # Example provider configurations
```

### Generated Files (Do Not Edit)
- All files in `internal/provider/` (except validators)
- All files in `internal/sdk/`
- Documentation in `docs/resources/` and `docs/data-sources/`

### Manual Files (Safe to Edit)
- `schemas/overlay.yaml` - Primary customization file
- `schemas/seqera-final.yaml` - OpenAPI spec with annotations
- `internal/validators/` - Some files have custom validation logic and are safe to edit
- `docs/internal/` - Internal documentation
- `examples/` - Example configurations

## Resource Implementation Pattern

Each Terraform resource follows this pattern:

1. **Resource Struct**: Defines the resource with SDK client
2. **Model Struct**: Defines the Terraform schema fields
3. **Schema Method**: Defines the resource schema with attributes
4. **CRUD Methods**: Create, Read, Update, Delete operations
5. **Type Conversion**: Methods to convert between Terraform and SDK types

## Customization Through Overlays

The `schemas/overlay.yaml` file is the primary mechanism for customizing the generated provider. 

## Build and Development Workflow

1. **Modify OpenAPI spec**: Update `seqera-final.yaml` with Speakeasy annotations
2. **Generate overlay**: Create overlay from changes using `speakeasy overlay compare`
3. **Test generation**: Run `speakeasy run --skip-versioning` to test changes
4. **Local testing**: Use examples in `examples/tests` with local provider builds
5. **Validate**: Run `speakeasy lint openapi -s seqera-final.yaml` for schema validation


## Testing and Validation

- **Local Development**: Use `examples/hello-world-aws` directory for integration testing, each cloud provider will have its own test folder. 
- **Provider Testing**: Test with `terraform plan` (avoid `terraform apply` during development), terraform init will not work with local providers. 
- **Schema Validation**: Use Speakeasy lint tools for OpenAPI validation
- **Generated Code**: Validate with `go build` and `go test`

## Important Notes

- **Generated Code**: Never manually edit generated files
- **Overlay First**: Always use overlays for customizations
- **Test Locally**: Use local provider builds for testing changes
