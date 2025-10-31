# Basic usage - fetch all data links in a workspace
data "seqera_data_links" "all" {
  workspace_id = seqera_workspace.my_workspace.id
}

# Create a map indexed by data link name for easy lookup
locals {
  data_links = {
    for dl in data.seqera_data_links.all.data_links : dl.name => dl
  }
}

# Access a specific data link by name
output "s3_bucket_url" {
  value = local.data_links["production-s3-bucket"].resource_ref
}

output "s3_bucket_region" {
  value = local.data_links["production-s3-bucket"].region
}

# Filter data links by provider
locals {
  aws_data_links = {
    for dl in data.seqera_data_links.all.data_links : dl.name => dl
    if dl.provider == "aws"
  }

  azure_data_links = {
    for dl in data.seqera_data_links.all.data_links : dl.name => dl
    if dl.provider == "azure"
  }
}

# Filter by provider AND region
locals {
  aws_us_east_1_data_links = {
    for dl in data.seqera_data_links.all.data_links : dl.name => dl
    if dl.provider == "aws" && dl.region == "us-east-1"
  }
}

# Use data links in resources
resource "seqera_pipeline" "example" {
  name         = "my-pipeline"
  workspace_id = seqera_workspace.my_workspace.id

  launch {
    work_dir = local.data_links["production-s3-bucket"].resource_ref
    # ... other configuration
  }
}
