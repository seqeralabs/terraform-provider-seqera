# Seqera Terraform Provider Release Guide

This document outlines the process for creating a new release of the Seqera Terraform Provider.

## Overview

The Seqera Terraform Provider is auto-generated using Speakeasy from OpenAPI specifications. Releases should be created when significant features, fixes, or improvements have been made to the provider.

## Pre-Release Checklist

Before creating a release, ensure the following items are completed:

### ✅ Version Management
- [ ] Version number updated in `.speakeasy/gen.yaml` (terraform.version field) and matches the provider version being released.
- [ ] Provider source name is correct in `gen.yaml` (author field)
- [ ] Ensure the Github release version is matching the Platform version being targeted.

## Release Process ( Master branch)

### 1. Ensure there are no changes that have not been commited.
```bash
# Ensure clean working directory
git status

# Update version in gen.yaml if not already done. This should match what the current Github release will be.
# Regenerate provider with final changes
speakeasy run --skip-versioning

```

### 2. Create Release Tag
```bash
# Create and push tag - Make sure to update the version to match what you have.
git tag -a v0.25.3
git push origin v0.25.3
```

### 3. Update the release notes

1. The release action will run and generate a new release in Github releases, matching the tag version.
2. The below template can be copy and pasted into the release.
3. Ensure you select generate release notes before saving the release.

## Release Notes Template

```markdown
## Seqera Terraform Provider v0.X.Y

[Brief description of release - new features, fixes, etc.]

### 📖 Documentation
- [Provider Documentation](https://github.com/seqeralabs/terraform-provider-seqera/tree/v0.X.Y/docs)
- [Examples](https://github.com/seqeralabs/terraform-provider-seqera/tree/v0.X.Y/examples)
- [Known Issues](https://github.com/seqeralabs/terraform-provider-seqera/blob/v0.X.Y/KNOWN-ISSUES.md)
- [Compatibility Matrix](https://github.com/seqeralabs/terraform-provider-seqera/blob/v0.X.Y/docs/compatibility-matrix.md)

### 🔧 Requirements
- Terraform >= 1.0
- Go >= 1.19 (for development)

### ✨ What's New
- [List new features and improvements]

### 🐛 Bug Fixes
- [List bug fixes]

### 💥 Breaking Changes (if any)
- [List breaking changes with migration guidance]

### 📋 Import Support Status
[Current status of import functionality for resources]

### ⚠️ Known Limitations
See [KNOWN-ISSUES.md](https://github.com/seqeralabs/terraform-provider-seqera/blob/v0.X.Y/KNOWN-ISSUES.md) for detailed information.
```

## Post-Release Actions

### Immediate
- [ ] Verify release appears correctly on GitHub
- [ ] Verify release on Hashicorp registry
