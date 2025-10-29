# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for the Seqera Platform API, generated using Speakeasy. The provider enables management of Seqera Platform resources through Terraform configurations.

**Key Architecture Points:**
- Auto-generated codebase using Speakeasy from OpenAPI specifications
- Go-based Terraform provider using terraform-plugin-framework
- Manual changes to internal files will be overwritten on next generation
- Provider supports resources and data sources for Seqera Platform entities

## Development Commands

### Building the Provider
```bash
# Build the provider binary
go build -o terraform-provider-seqera

# Build and run with debug mode for local development
To test run the terraform in ./examples/tests
```

### Local Development and Testing
```bash
cd examples/tests
TF_REATTACH_PROVIDERS=... terraform init
TF_REATTACH_PROVIDERS=... terraform apply

# Alternative: Use compiled provider with .terraformrc dev_overrides
go build
# Configure ~/.terraformrc with dev_overrides pointing to the binary path
```

### Code Generation
```bash

# Create a new overaly file from the update openAPI spec (example). The file must be name seqera-final.yaml for speakeasy to pick it up.
speakeasy overlay compare --before=seqera-api-latest.yml --after=seqera-final.yaml > overlay_new.yaml
# Regenerate provider code using Speakeasy
speakeasy run --skip-versioning 

```

## Project Structure

### Generated Core Provider Code 
- `internal/provider/` - Main provider implementation
  - `provider.go` - Core provider configuration
  - `*_resource.go` - Resource implementations
  - `*_data_source.go` - Data source implementations
  - `*_sdk.go` - SDK integration layers
  - `types/` - Terraform schema type definitions
  - `reflect/` - Reflection utilities for type conversion
  - `validators/` - Custom validation logic

### Generated SDK
- `internal/sdk/` - Auto-generated SDK for Seqera API
  - API client implementations for all endpoints
  - Model definitions in `models/shared/`
  - HTTP client configuration and utilities

### Configuration
- `.speakeasy/` - Speakeasy configuration and generation artifacts
  - `gen.yaml` - Generation configuration
  - `workflow.yaml` - Workflow definition
  - `out.openapi.yaml` - Processed OpenAPI specification
- `overlays/` - Speakeasy overlay files for customizing resource generation
- `schemas/` - OpenAPI specifications

### Documentation and Examples
- `docs/` - Generated Terraform provider documentation
- `examples/` - Example Terraform configurations for testing
- `examples/tests/` - Test configurations
- `.genignore` - Files to protect from Speakeasy regeneration
- `internal-docs/` - Internal development guides
  - `OVERLAY_GUIDE.md` - Best practices for overlays, validators, and resource documentation
  - `CREDS_HOISTING_GUIDE.md` - Guide for implementing credential hoisting
  - `SPEAKEASY_EXTENSIONS_REFERENCE.md` - Complete Speakeasy extensions reference

## Available Resources and Data Sources

### Resources
- `seqera_action` - Seqera actions/workflows
- `seqera_compute_env` - Compute environments 
- `seqera_credential` - Authentication credentials
- `seqera_data_studios` - Data studio instances
- `seqera_orgs` - Organizations
- `seqera_pipeline` - Pipeline definitions
- `seqera_tokens` - Access tokens

### Data Sources
- Multiple data sources for listing and querying existing resources
- Single item data sources for specific resource lookup
- User and workspace data sources for account information

## Development Guidelines

### Overlay and Resource Best Practices
**IMPORTANT**: When working with overlays, resource documentation, or custom validators, always consult `internal-docs/OVERLAY_GUIDE.md` for:
- Overlay file structure and organization patterns
- Field cleanup guidelines (removing unmanageable/internal fields)
- Resource example best practices
- Custom validator patterns and implementation
- Documentation verification steps

### Code Generation Workflow
1. Only modify the OpenAPI specifications in `schemas/seqera-final.yaml` to add speakeasy annotations.
2. Generate the overlay file from the edited specification using `speakeasy overlay compare --before=seqera-api-latest.yml --after=seqera-final.yaml > overlay_new.yaml`.
3. Edit overlay files in `overlays/` directory following patterns in `internal-docs/OVERLAY_GUIDE.md`
4. Create custom examples in `examples/resources/seqera_[resource]/resource.tf` and protect them with `.genignore`
5. Run `speakeasy run --skip-versioning` to regenerate provider code
6. Verify generated documentation in `docs/resources/[resource].md`
7. Test changes with local provider builds: `go build -o terraform-provider-seqera`
8. Test with `terraform plan` in `examples/tests/` (do not apply)
9. You can run `speakeasy lint openapi -s seqera-final.yaml` to check for schema errors

### Testing
- Use `examples/tests` directory for integration testing
- Test both resource creation and data source queries
- Verify provider behavior with `terraform plan`

### Authentication
Provider supports multiple authentication methods configured through:
- Environment variables
- Provider configuration block
- OAuth2 client credentials and password flows

## Important Notes

- **Generated Code**: Most files are auto-generated - manual changes will be lost
- **Contributions**: Report issues via GitHub issues rather than direct PRs
- **Speakeasy Integration**: Uses Speakeasy for OpenAPI-to-Terraform generation
- **Terraform Framework**: Built on terraform-plugin-framework (not legacy SDK)