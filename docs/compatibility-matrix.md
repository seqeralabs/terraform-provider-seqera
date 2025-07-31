# Compatibility Matrix

This document shows the compatibility between Seqera platform versions, Seqera platform API versions, and Terraform provider versions.

## Version Compatibility

| Platform Type | Platform Version | API Version | Terraform Provider Version | Notes |
|---------------|------------------|-------------|----------------------------|-------|
| **Seqera Platform** | Latest | 1.56.0 | 0.25.0  | Current cloud deployment and Latest enterprise release |
| **Seqera Platform** | v23.4 | 1.56.0 | 0.23.0 | Enterprise release |
| **Seqera Platform** | v24.2 | 1.45.0 | 0.24.0  | Enterprise release |

## Additional Information

When using the Terraform provider:

1. Ensure your Seqera platform version matches one of the supported configurations above. 
2. For enterprise deployments, verify your platform version against the supported list. Latest implies the latest available enterprise version. 
3. For Seqera cloud users, the cloud deployments are always using the latest API specification and platform version. 

For issues or questions about compatibility, please refer to the [troubleshooting documentation](internal/troubleshooting.md).


:::note The API and Terraform provider uses the semantic versioning convention (major.minor.patch). In the event that a breaking change is introduced in future versions, we will publish guidance on the v1 support schedule and steps to mitigate disruption to your production environment. The following do not constitute breaking changes:

Adding new API endpoints, new HTTP methods to existing endpoints, request parameters, or response fields
Adding new values to existing enums or string constants
Expanding accepted input formats or value ranges
Adding new optional headers or query parameters
Improving error messages or adding new error codes
Deprecation warnings (without removal)
Clients should be designed to gracefully handle unknown enum values, ignore unrecognized response fields, and not rely on specific error message text. :::

