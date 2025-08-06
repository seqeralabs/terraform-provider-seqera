# Testing the provider in Dev

When working on the provider locally, terraform will try and pull the provider in the terraform registry. To prevent this from ocurring we will need to update the .terraformrc file in your $HOME directory to point to your local provider.

***NOTE***
If you have this setting enabled, this will also override your ability to work with the external provider in the terraform registry.

1. Update you .terraformrc to point to your provider:

Example `.terraformrc`:
```sh
provider_installation {

  dev_overrides {
  "registry.terraform.io/speakeasy/seqera" = "/Users/shahbaz.mahmood/workspace/devops/terraform-provider-seqera"
  }

  direct {}
}
```


2. Ensure the provider in your terraform has the same source, as what is specified in the `.terraformrc` :

```hcl
terraform {
  required_providers {
    seqera = {
      source  = "registry.terraform.io/speakeasy/seqera"
      #version = "0.0.3"
    }
  }
}
```

3. Build the provider locally, so the binary is available, you will need to build the repo using the below command:

```sh
shahbaz.mahmood@Shahbazs-MacBook-Pro terraform-provider-seqera % go build .
```


4. Declare an environment variables with your bearer token and optionally the server URL for enterprise of dev instances:
```sh
export TF_VAR_seqera_server_url="https://api.cloud.seqera.io"
export TF_VAR_seqera_bearer_auth="your-token-here"
```


## Using this Terraform.

Currently all terraform state is stored locally on your machine. Additionally for each cloud provider we need to provide some enviornment variables for access keys and such. List of required environment variables:
```sh
TF_VAR_seqera_server_url=https://api.cloud.seqera.io # or enterprise URL
TF_VAR_seqera_bearer_auth=$BEARER_KEY
TF_VAR_secret_key=$AWS_SECRET_KEY
TF_VAR_access_key=$AWS_ACCESS_KEY
TF_VAR_azure_batch_key=$AZURE_BATCH_KEY
TF_VAR_azure_storage_key=$AZURE_STORAGE_KEY
```

Additionally, you will need to update the locals terraform code with your working directories for each cloud provider.
```terraform
locals {
  service_account_key = file("${path.module}/service-account-key.json")
  gcp_work_dir = "gs://terraform-provider-testing"
  azure_batch_name = "seqeralabs"
  azure_storage_name = "seqeralabs"
  azure_work_dir = "az://terraform-provider"
}
```

Finally, for GCP make sure you service account key json is present in the below location:
`service_account_key = file("${path.module}/service-account-key.json")`


***NOTE***
You may need to run `terraform init` for the module calls to work, the command will fail but you should be able to run tf apply without failure.
