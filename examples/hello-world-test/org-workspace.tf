data "seqera_user" "my_user" {
  user_id = 7672
}

data "seqera_user_workspaces" "test" {
  user_id = 7672
}

data "seqera_current_user" "my_currentuser" {
}



# output "user_id" {
#   value = data.seqera_user.my_user.user_id
# }


# output "current_user_id" {
#   value = data.seqera_current_user.my_currentuser
# }
# output "test_workspace" {
#   value = data.seqera_user_workspaces.test
# }


# data "seqera_orgs" "my_orgs" {
# }

# output "orgs" {
#   value = data.seqera_orgs.my_orgs
  
# }


resource "seqera_orgs" "test_org" {
  description = "testing org for the terraform provider"
  full_name   = "seqera_test_shahbaz_tf_provider"
  name        = "seqera_test_shahbaz_tf_provider"
}

resource  "seqera_workspace" "my_workspace" {
  description = "A test workspace created with Terraform"
  name        = "test-workspace-tf"
  full_name   = "Test Workspace for Terraform Provider"
  org_id     = resource.seqera_orgs.test_org.org_id
  visibility = "PRIVATE"
}

# resource "seqera_tokens" "my_tokens" {
#   name     = "terraform-test-token"
# }

