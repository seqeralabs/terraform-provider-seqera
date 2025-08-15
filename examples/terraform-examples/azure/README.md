# Azure Example - Seqera Platform Terraform Provider

This example demonstrates a complete Azure setup for running bioinformatics workflows using the Seqera Platform. It includes organization setup, Azure Batch compute environment, and a configured nf-core/rnaseq pipeline.

## What This Example Includes

- **Organization and Workspace**: Multi-tenant structure optimized for Azure workflows
- **Azure Credentials**: Azure Batch and Storage account integration
- **Compute Environment**: Azure Batch with configurable VM sizing and networking
- **Pipeline & Workflow**: nf-core/rnaseq v3.19.0 with test profile for analysis
- **Data Management**: Azure Storage integration with datasets
- **Security Features**: Pipeline secrets and workspace labels
- **External Scripts**: Pre and post-run bash scripts for workflow customization

> [!WARNING]
> This will create cloud resources that will incur costs. Please be aware of your current cloud spend and remove resources when you are done.

## How to Run This Example

1. **Navigate to the Azure example directory**:

   ```bash
   cd examples/terraform-examples/azure
   ```

2. **Copy the example variables file**:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

3. **Edit the terraform.tfvars file** with your Azure credentials and settings:

   ```bash
   # Required Azure credentials
   batch_key    = "your-azure-batch-access-key"
   batch_name   = "your-azure-batch-account-name"
   storage_key  = "your-azure-storage-access-key"
   storage_name = "your-azure-storage-account-name"
   
   # Azure region
   azure_region = "East US"  # or your preferred region
   
   # Working directory (Azure Blob Storage path)
   work_dir = "az://your-container/work"
   ```

4. **Initialize Terraform**:

   ```bash
   terraform init
   ```

5. **Review the execution plan**:

   ```bash
   terraform plan
   ```

6. **Apply the configuration**:

   ```bash
   terraform apply
   ```

## Important Notes

⚠️ **Organization Name**: You must update the organization name in `main.tf` to be unique. The default `azure-example-org` may already exist. Change it to something like `your-company-azure-org`.

```hcl
resource "seqera_orgs" "my_org" {
  description = "Example org for running Azure centric workflows"
  full_name   = "your-company-azure-org"  # Change this to be unique
  name        = "your-company-azure-org"  # Change this to be unique
}
```

## Prerequisites

- Azure account with appropriate permissions for Batch, Storage, and Compute resources
- Azure Batch account with access keys
- Azure Storage account with access keys
- Terraform >= 1.0
- Valid Seqera Platform account and API credentials

## Resources Created

This example creates the following resources:

- Seqera Organization
- Seqera Workspace
- Azure Credentials
- Azure Batch Compute Environment
- Pipeline Secret
- Pipeline (nf-core/rnaseq)
- Workflow
- Labels
- Dataset

## Cleanup

To destroy all created resources:

```bash
terraform destroy
```

## Customization

You can customize this example by:

- Modifying VM sizes in the compute environment forge configuration
- Changing the pipeline to a different nf-core workflow
- Adjusting resource configurations in `variables.tf`
- Updating the pre/post-run scripts in the `scripts/` directory
- Configuring additional Azure-specific networking or security settings
