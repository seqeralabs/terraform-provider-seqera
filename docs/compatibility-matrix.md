# Compatibility Matrix

This document shows the compatibility between Seqera platform versions, Seqera platform API versions, and Terraform provider versions.

## Version Compatibility

| Platform Type | Platform Version | API Version | Terraform Provider Version | Notes |
|---------------|------------------|-------------|----------------------------|-------|
| **Seqera Cloud** | Latest | 1.56.0 | 1.0+ | Current cloud deployment |
| **Seqera Enterprise** | v23.4 | 1.56.0 | 1.0+ | Enterprise release |
| **Seqera Enterprise** | v24.2 | 1.45.0 | 0.9 | Enterprise release |
| **Seqera Enterprise** | v25.1 | 1.56.0 | 1.0+ | Latest enterprise release |

## Support Information

- **Terraform Provider v1.0+**: Supports all listed platform versions through their respective API specifications
- **API Specifications**: Each platform version uses its own OpenAPI specification file located in the `specs/` directory
- **Backward Compatibility**: The Terraform provider is designed to work with multiple API versions simultaneously through specification-driven generation

## Configuration Notes

When using the Terraform provider:

1. Ensure your Seqera platform version matches one of the supported configurations above
2. For enterprise deployments, verify your platform version against the supported list
3. Cloud deployments always use the latest API specification

For issues or questions about compatibility, please refer to the [troubleshooting documentation](internal/troubleshooting.md).
