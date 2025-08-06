# Compatibility Matrix

This document shows the compatibility between Seqera platform versions, Seqera platform API versions, and Terraform provider versions.

## Version Compatibility

| Platform Type | Platform Version | API Version | Terraform Provider Version |
|---------------|------------------|-------------|----------------------------|
| **Seqera Platform** | Latest | 1.56.0 | 0.25.0  |
| **Seqera Platform** | v23.4 | 1.56.0 | 0.23.0 |
| **Seqera Platform** | v24.2 | 1.45.0 | 0.24.0  |

## Additional Information

When using the Terraform provider:

1. Ensure your Seqera platform version matches one of the supported configurations above.
2. For enterprise deployments, verify your platform version against the supported list. Latest implies the latest available enterprise version.
3. For Seqera cloud users, the cloud deployments are always using the latest API specification and platform version.

For issues or questions about compatibility, please refer to the [troubleshooting documentation](internal/troubleshooting.md).

## ⚠️ Important: Semantic Versioning and Breaking Changes

The API and Terraform provider uses the semantic versioning convention (major.minor.patch). In the event that a breaking change is introduced in future versions, we will publish guidance on the v1 support schedule and steps to mitigate disruption to your production environment.

### The following do NOT constitute breaking changes:

* Adding new resources or data sources
* Adding new optional attributes to existing resources or data sources
* Adding new values to existing enum attributes or string constants
* Expanding accepted input formats or value ranges for attributes
* Adding new optional provider configuration parameters
* Improving error messages or adding new error codes
* Adding new computed attributes to existing resources
* Adding new import capabilities to existing resources
* Deprecation warnings (without removal)
* Bug fixes that don't change the resource schema
* Performance improvements to provider operations

### Best Practices

Terraform configurations should be designed to gracefully handle new optional attributes and not rely on specific error message text. Use of `lifecycle` blocks with `ignore_changes` can help manage unexpected attribute additions during upgrades. Always test provider upgrades in non-production environments first.
