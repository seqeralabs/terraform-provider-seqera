# Internal Documentation

This directory contains internal development guides for maintaining and extending the Seqera Terraform Provider.

## Contents

### [OVERLAY_GUIDE.md](./OVERLAY_GUIDE.md)
Best practices for creating and maintaining Speakeasy overlay files. Covers:
- Overlay file structure and organization
- Field cleanup guidelines
- Resource example patterns
- Custom validator implementation
- Documentation verification

### [SPEAKEASY_EXTENSIONS_REFERENCE.md](./SPEAKEASY_EXTENSIONS_REFERENCE.md)
Complete reference of Speakeasy OpenAPI extensions for SDK and Terraform provider generation. Covers:
- General SDK extensions (naming, enums, documentation, runtime behavior)
- Terraform-specific extensions (resource mapping, constraints, validation, state management)
- Usage notes and common patterns
- Important warnings about non-existent or deprecated extensions

## Audience

These guides are for:
- Provider maintainers
- Contributors adding new resources
- Developers debugging provider generation issues

## User-Facing Documentation

User-facing documentation is located in:
- `/docs/` - Generated Terraform provider documentation
- `/README.md` - Provider overview and getting started
- `/USAGE.md` - Usage examples
