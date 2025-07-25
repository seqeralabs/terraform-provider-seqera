# AWS Example - Seqera Platform Terraform Provider

This example demonstrates a complete AWS setup for running bioinformatics workflows using Seqera Platform. It includes organization setup, AWS Batch compute environment, and a configured nf-core/rnaseq pipeline. 

## What This Example Includes

- **Organization and Workspace**: Multi-tenant structure for organizing workflows
- **AWS Credentials**: IAM integration with support for assume roles
- **Compute Environment**: AWS Batch with configurable EC2 instances, Fusion2, and Wave integration
- **Pipeline & Workflow**: nf-core/rnaseq v3.19.0 with test profile for analysis
- **Data Management**: S3 integration with datasets and data links
- **Security Features**: Pipeline secrets and workspace labels
- **External Scripts**: Pre and post-run bash scripts for workflow customization

## How to Run This Example

1. **Navigate to the AWS example directory**:
   ```bash
   cd examples/terraform-examples/aws
   ```

2. **Copy the example variables file**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

3. **Edit the terraform.tfvars file** with your AWS credentials and settings:
   ```bash
   # Required AWS credentials
   access_key = "your-aws-access-key"
   secret_key = "your-aws-secret-key"
   aws_region = "us-east-1"  # or your preferred region
   
   # Working directory (S3 bucket path)
   work_dir = "s3://your-bucket/work"
   
   # Optional: IAM role for cross-account access
   iam_role = "arn:aws:iam::123456789012:role/SeqeraRole"  # optional
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

⚠️ **Organization Name**: You must update the organization name in `main.tf` to be unique. The default `aws-example-org` may already exist. Change it to something like `your-company-aws-org`.

```hcl
resource "seqera_orgs" "my_org" {
  description = "Example org for running AWS centric workflows"
  full_name   = "your-company-aws-org"  # Change this to be unique
  name        = "your-company-aws-org"  # Change this to be unique
}
```

## Prerequisites

- AWS account with appropriate permissions for Batch, EC2, S3, and IAM
- Terraform >= 1.0
- Valid Seqera Platform account and API credentials

## Resources Created

This example creates the following resources:
- Seqera Organization
- Seqera Workspace  
- AWS Credentials
- AWS Batch Compute Environment
- Pipeline Secret
- Data Link (S3)
- Action
- Pipeline (nf-core/rnaseq)
- Workflow
- Data Studio
- Labels
- Teams
- Dataset

## Cleanup

To destroy all created resources:

```bash
terraform destroy
```

## Customization

You can customize this example by:
- Modifying instance types in the compute environment
- Changing the pipeline to a different nf-core workflow
- Adjusting resource configurations in `variables.tf`
- Updating the pre/post-run scripts in the `scripts/` directory