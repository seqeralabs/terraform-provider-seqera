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

### [STATE_UPGRADER_GUIDE.md](./STATE_UPGRADER_GUIDE.md)
The **default** pattern for writing state upgraders so prior state upgrades
cleanly (no "unsupported attribute" errors when the schema drops fields). Covers:
- When a schema version bump / upgrader is actually needed (rarely)
- Why hand-stripping removed attributes is wrong, and the framework mechanism behind it
- The inject-schema + lenient-re-decode pattern (since Speakeasy can't emit `PriorSchema`)
- Testing against the real schema and end-to-end verification

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
