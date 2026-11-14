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

# Run the Go unit tests (validators, hooks, provider helpers)
go test ./internal/...
```

### Local Development and Testing
```bash
# Use the compiled provider through ~/.terraformrc dev_overrides pointing at
# this repository directory (no `terraform init` needed with dev_overrides)
go build -o terraform-provider-seqera

# Write a throwaway configuration in a scratch directory, then:
terraform validate   # runs schema + plan validators, no API access needed
terraform plan       # needs Seqera credentials (do not apply)

# Alternative: debug mode with TF_REATTACH_PROVIDERS
```

### Code Generation
```bash
# Regenerate provider code from specs/seqera-api-cloud.yaml + overlays/
speakeasy run --skip-versioning
```

The pipeline is declared in `.speakeasy/workflow.yaml`: the upstream spec
`specs/seqera-api-cloud.yaml` is loaded, every file in `overlays/` is applied in
the listed order, and the result is written to `.speakeasy/out.openapi.yaml`
before generation. To update the upstream spec, follow
`docs-internal/SPEC_UPDATE_RUNBOOK.md`.

## Project Structure

### Generated Core Provider Code 
- `internal/provider/` - Main provider implementation
  - `provider.go` - Core provider configuration
  - `*_resource.go` - Resource implementations
  - `*_data_source.go` - Data source implementations
  - `*_sdk.go` - SDK integration layers
  - `types/` - Terraform schema type definitions
  - `reflect/` - Reflection utilities for type conversion
- `internal/validators/` - Custom plan validators (hand-written, listed in `.genignore`)

### Generated SDK
- `internal/sdk/` - Auto-generated SDK for Seqera API
  - API client implementations for all endpoints
  - Model definitions in `models/shared/`
  - HTTP client configuration and utilities

### Configuration
- `.speakeasy/` - Speakeasy configuration and generation artifacts
  - `gen.yaml` - Generation configuration
  - `workflow.yaml` - Workflow definition
  - `out.openapi.yaml` - Processed OpenAPI specification (spec + overlays; read-only, regenerated)
- `overlays/` - Speakeasy overlay files: **all** Speakeasy annotations and customisations live here
- `specs/seqera-api-cloud.yaml` - Upstream Seqera Platform OpenAPI spec, kept free of `x-speakeasy-*` annotations

### Documentation and Examples
- `docs/` - Generated Terraform provider documentation
- `examples/resources/`, `examples/data-sources/`, `examples/provider/` - Examples embedded in the docs
- `examples/terraform-examples/` - Full example configurations per cloud (aws, azure, gcp)
- `.genignore` - Files to protect from Speakeasy regeneration
- `docs-internal/` - Internal development guides
  - `OVERLAY_GUIDE.md` - Best practices for overlays, validators, resource documentation, and nested-structure hoisting (transform-based + legacy patterns)
  - `SPEAKEASY_EXTENSIONS_REFERENCE.md` - Complete Speakeasy extensions reference
  - `SPEC_UPDATE_RUNBOOK.md` - **Guide to updating the API spec**; fetching, reviewing, and validating changes before they reach Terraform
  - `STATE_UPGRADER_GUIDE.md` - **Default pattern for state upgraders** (inject schema + lenient re-decode); consult before adding/editing any `*_resource` state upgrader or bumping `x-speakeasy-entity-version`

## Available Resources and Data Sources

### Resources
See `docs/resources/` for the full list. Main families:
- Compute environments: `seqera_compute_env` (generic) and per-platform resources such as `seqera_aws_batch_ce`, `seqera_gcp_batch_ce`, `seqera_azure_batch_ce`
- Credentials: `seqera_credential` (generic) and per-provider resources such as `seqera_aws_credential`
- `seqera_pipeline`, `seqera_action`, `seqera_pipeline_action` - Pipelines and launch actions
- `seqera_studios` - Studio instances
- `seqera_orgs`, `seqera_teams`, `seqera_workspace` - Organisation structure
- `seqera_tokens`, `seqera_service_account` - Access tokens and service accounts

### Data Sources
- Multiple data sources for listing and querying existing resources
- Single item data sources for specific resource lookup
- User and workspace data sources for account information

## Development Guidelines

### Overlay and Resource Best Practices
**IMPORTANT**: When working with overlays, resource documentation, or custom validators, always consult `docs-internal/OVERLAY_GUIDE.md` for:
- Overlay file structure and organization patterns
- Field cleanup guidelines (removing unmanageable/internal fields)
- Resource example best practices
- Custom validator patterns and implementation
- Documentation verification steps

### Code Generation Workflow
1. Never add Speakeasy annotations to `specs/seqera-api-cloud.yaml`; it must stay identical to the upstream spec.
2. Make every customisation in the matching file under `overlays/` (e.g. `overlays/studios.yaml`, `overlays/compute-env.yaml`), following `docs-internal/OVERLAY_GUIDE.md`. A new overlay file must also be added to `.speakeasy/workflow.yaml`.
   - Hide a server-managed or caller-dependent field with `remove: true` on its response schema property (e.g. `$.components.schemas.DataStudioDto.properties.<field>`).
   - Attach validators with `x-speakeasy-plan-validators: <Name>` (a single name or a list).
3. Create custom examples in `examples/resources/seqera_[resource]/resource.tf` and protect them with `.genignore`
4. Run `speakeasy run --skip-versioning` to regenerate provider code
5. Verify generated documentation in `docs/resources/[resource].md`
6. Test changes with local provider builds: `go build -o terraform-provider-seqera` and `go test ./internal/...`
7. Test with `terraform validate` / `terraform plan` against a scratch configuration (do not apply)
8. You can run `speakeasy lint openapi -s .speakeasy/out.openapi.yaml` to check the processed spec for errors

### Custom Validators
- Speakeasy scaffolds one file per validator name: `x-speakeasy-plan-validators: GlobalCachePathS3Validator` on a string field expects `internal/validators/stringvalidators/global_cache_path_s3_validator.go` exporting `GlobalCachePathS3Validator()`. If that file is missing, a TODO stub is generated that fails every plan; if the function is defined in a differently named file, compilation fails with a redeclaration.
- Put logic shared by several validators in a file whose name does not match any validator name (e.g. `global_cache_path_common.go`) and keep the per-name files as thin constructors.
- Add every hand-written validator and its test to `.genignore`.
- Validators that need sibling fields read them with `req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("<sibling>"), ...)`; always allow unknown values.

### Testing
- Unit-test validators with `tfsdk.Config` fixtures (see `internal/validators/stringvalidators/*_test.go`)
- Use `terraform validate` with the dev_overrides build to exercise plan validators without credentials
- Test both resource creation and data source queries
- Verify provider behavior with `terraform plan`

### Authentication
Provider supports two authentication methods, each configurable through the provider block or environment variables:
- Personal access token: `bearer_auth` / `TOWER_ACCESS_TOKEN`
- Service account OAuth2 client credentials: `client_id` / `TOWER_CLIENT_ID` and `client_secret` / `TOWER_CLIENT_SECRET`, exchanged at `{server_url}/token`. The scheme is declared in `overlays/auth.yaml` because the platform hides the token endpoint from its published spec.

## Important Notes

- **Generated Code**: Most files are auto-generated - manual changes will be lost
- **Contributions**: Report issues via GitHub issues rather than direct PRs
- **Speakeasy Integration**: Uses Speakeasy for OpenAPI-to-Terraform generation
- **Terraform Framework**: Built on terraform-plugin-framework (not legacy SDK)
