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


# resource "seqera_tokens" "my_tokens" {
#   name     = "terraform-test-token"
# }

