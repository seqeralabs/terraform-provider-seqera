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
