# GCP Example - Seqera Platform Terraform Provider

This example demonstrates a complete Google Cloud Platform setup for running bioinformatics workflows using the Seqera Platform. It includes organization setup, Google Batch compute environment optimized for genomics workloads, and a configured nf-core/rnaseq pipeline.

## What This Example Includes

- **Organization and Workspace**: Multi-tenant structure for GCP workflows
- **GCP Credentials**: Service account key integration for secure access
- **Compute Environment**: Google Batch with n2-standard-8 instances (8 vCPUs, 32GB RAM) optimized for genomics
- **Performance Features**: Fusion2 and Wave enabled, spot instances for cost optimization
- **Pipeline & Workflow**: nf-core/rnaseq v3.19.0 with test profile for genomics analysis
- **Data Management**: Google Cloud Storage integration with datasets
- **Security Features**: Pipeline secrets and workspace labels
- **External Scripts**: Pre and post-run bash scripts for workflow customization

> [!WARNING]
> This will create cloud resources that will incur costs. Please be aware of your current cloud spend and remove resources when you are done.

## How to Run This Example

1. **Navigate to the GCP example directory**:

   ```bash
   cd examples/terraform-examples/gcp
   ```

2. **Copy the example variables file**:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

3. **Edit the terraform.tfvars file** with your GCP settings:

   ```bash
   # Required GCP settings
   gcp_region = "us-central1"  # or your preferred region
   gcp_location = "us-central1-a"  # or your preferred zone
   
   # Service account key file path
   service_account_key = "/path/to/your/service-account-key.json"
   
   # Working directory (Google Cloud Storage bucket path)
   work_dir = "gs://your-bucket/work"
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

⚠️ **Organization Name**: You must update the organization name in `main.tf` to be unique. The default `gcp-example-org` may already exist. Change it to something like `your-company-gcp-org`.

```hcl
resource "seqera_orgs" "my_org" {
  description = "Example org for running GCP centric workflows"
  full_name   = "your-company-gcp-org"  # Change this to be unique
  name        = "your-company-gcp-org"  # Change this to be unique
}
```

## Prerequisites

- Google Cloud Platform account with appropriate permissions
- Service account with necessary roles:
  - Batch Job Editor
  - Compute Instance Admin
  - Storage Admin
  - Service Account User
- Service account key file downloaded locally
- Terraform >= 1.0
- Valid Seqera Platform account and API credentials

## Compute Configuration

This example uses **n2-standard-8** instances (8 vCPUs, 32GB RAM) which are optimized for genomics workloads like nf-core/rnaseq. The configuration includes:

- **Machine Type**: n2-standard-8 for processing large genomic datasets
- **Head Job**: 4 vCPUs, 16GB RAM for workflow orchestration
- **Performance**: Fusion2 and Wave enabled for faster data access
- **Cost Optimization**: Spot instances enabled for reduced compute costs
- **Storage**: 100GB boot disk for large dataset processing

## Resources Created

This example creates the following resources:

- Seqera Organization
- Seqera Workspace
- GCP Credentials
- Google Batch Compute Environment
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

- Modifying machine types in the compute environment (e.g., n2-highmem-8 for memory-intensive workflows)
- Changing the pipeline to a different nf-core workflow
- Adjusting resource configurations in `variables.tf`
- Updating the pre/post-run scripts in the `scripts/` directory
- Configuring additional GCP-specific settings like custom networks or disk types
