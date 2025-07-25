# Adding New Resources to the Seqera Platform Terraform Provider

This guide explains how to add new resources to the Seqera Platform Terraform provider, which is generated using Speakeasy from OpenAPI specifications.

## References

- [Speakeasy Extensions Documentation](https://www.speakeasy.com/docs/speakeasy-reference/extensions)
- [Speakeasy Terraform Provider Creation](https://www.speakeasy.com/docs/create-terraform)

## How Speakeasy Generation Works

Speakeasy generates Terraform providers by analyzing OpenAPI specifications and creating corresponding resource and data source implementations. The generation process relies on special annotations in the OpenAPI schema to understand how to map API operations to Terraform CRUD operations.

The key distinction is between:
- **`x-speakeasy-entity`**: Applied at the schema level to define a Terraform resource entity
- **`x-speakeasy-entity-operation`**: Applied at the operation level to map specific HTTP operations to Terraform CRUD actions (create, read, update, delete)

This mapping allows Speakeasy to automatically generate the appropriate Terraform resource lifecycle methods from REST API endpoints.

## Adding a New Resource: DataLink Example

To add a new resource to the provider, you need to create an overlay that adds the necessary Speakeasy annotations to the OpenAPI specification. Here's a complete example for adding a DataLink resource:

```yaml
overlay: 1.0.0
info:
  title: Add DataLink Resource
  version: 0.0.1
actions:
  # Map CRUD operations to Terraform lifecycle methods
  - target: $["paths"]["/data-links"]["post"]
    update:
      x-speakeasy-entity-operation: DataLink#create
  
  - target: $["paths"]["/data-links/{dataLinkId}"]["get"]
    update:
      x-speakeasy-entity-operation: DataLink#read
  
  - target: $["paths"]["/data-links/{dataLinkId}"]["put"]
    update:
      x-speakeasy-entity-operation: DataLink#update
  
  - target: $["paths"]["/data-links/{dataLinkId}"]["delete"]
    update:
      x-speakeasy-entity-operation: DataLink#delete

  # Customize field names in generated code, this was needed to 
  - target: $["components"]["schemas"]["DataLinkDto"]["properties"]["id"]
    update:
      x-speakeasy-name-override: dataLinkId
  
  - target: $["components"]["schemas"]["DataLinkProvider"]
    update:
      x-speakeasy-name-override: providerType
```

### Key Components:

1. **CRUD Operation Mapping**: Each HTTP operation is mapped to a Terraform action using the format `ResourceName#action`
2. **Field Name Overrides**: Use `x-speakeasy-name-override` to customize how schema properties are named in the generated Terraform code
3. **Path Parameters**: Ensure path parameters match the expected Terraform resource identifier patterns

## Speakeasy Commands

### Applying Overlays

To apply an overlay to your OpenAPI schema:

```bash
speakeasy overlay apply --schema seqera-api-latest.yml --overlay overlay_new.yaml
```

### Generating Overlays

You can generate an overlay by comparing two OpenAPI specifications. This is useful when you've manually edited a schema and want to create a reusable overlay:

```bash
speakeasy overlay compare --before=seqera-api-latest.yml --after=seqera-final.yaml > overlay_new.yaml
```

> **Note**: As documented in CLAUDE.md, the final schema file must be named `seqera-final.yaml` for Speakeasy to pick it up during generation.

### Linting OpenAPI Specifications

Before generating code, it's important to validate your OpenAPI specification for errors and inconsistencies:

```bash
speakeasy lint openapi -s seqera-final.yaml
```

### Code Generation

After creating or updating overlays, regenerate the provider code:

***Note*** On occastion the run command may need to be run twice. Not entirely sure why this occurs but seems due to caching since the first try results in some linting error and the second try works. 

```bash
speakeasy run --skip-versioning # This skips the increasing the version of the provider, when doing actual releases we do not use this flag. 

```

## Documentation Generation

The provider documentation is automatically generated using `terraform-plugin-docs`. This tool reads the provider schema and generates markdown documentation for all resources and data sources. You can simply run the `speakeasy run` command to regenerate the docs. 

## Workflow Summary

1. **Identify API Endpoints & Schema**: Determine which API endpoints should map to Terraform CRUD operations
2. **Create Overlay**: Write an overlay file with appropriate `x-speakeasy-entity-operation` annotations
3. **Apply Overlay**: Use `speakeasy overlay apply` to modify the OpenAPI schema
4. **Generate Code**: Run `speakeasy run --skip-versioning` to generate the new provider code
5. **Test**: Build and test the provider with the new resource
6. **Generate Docs**: Run `terraform-plugin-docs generate` to create documentation

## Best Practices

- Use consistent naming patterns for resources and operations
- Ensure all CRUD operations are properly mapped
- Test the generated resource thoroughly with real API calls
- Validate that the Terraform state management works correctly
- Update examples and documentation to demonstrate the new resource

## Troubleshooting

- If resources aren't generated, verify that all required operations (at minimum create, read, delete) are mapped
- Check that path parameters match Terraform's expected identifier patterns
- Ensure the OpenAPI specification is valid before applying overlays
- Review generated code for any naming conflicts or unexpected transformations